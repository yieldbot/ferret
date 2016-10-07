/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

/* jslint browser: true */
/* global document: false, $: false, Rx: false, unescape: false */
'use strict';

// Create the module
var app = function app() {

  if(typeof $ != "function" || typeof Rx != "object") {
    throw new Error("missing or invalid libraries (jQuery, Rx)");
  }

  // Init vars
  var serverUrl    = location.protocol + '//' + location.hostname + ':' + location.port,
      appPath      = location.pathname || '/',
      providerList = [];

  // for debugging
  if(location.protocol == 'file:') {
    serverUrl = 'http://localhost:3030';
    appPath = '/';
  }

  // init initializes the app
  function init() {
    // Get providers
    providersGet().then(
      function(data) {
        if(data && data instanceof Array && data.length > 0) {
          // Sort the provider list base on priority
          providerList = data.sort(function(a, b) {
            return b.priority - a.priority;
          });
        } else {
          critical('There is no any available provider to search');
          return;
        }
        providersPrepare();
        listen();
        urlQueryHandler();
      },
      function(err) {
        var e = parseError(err);
        critical("Could not get the available providers due to " + e.message  + " (" + e.code + ")");
      }
    );
  }

  // listen listens search requests for the given providers
  function listen() {

    // Create the observables

    // Click button
    var clickSource = Rx.Observable
      .fromEvent($('#searchButton'), 'click')
      .map(function() { return $('#searchInput').val(); });

    // Input field
    var inputSource = Rx.Observable
      .fromEvent($('#searchInput'), 'keyup')
      .filter(function(e) { return (e.keyCode == 13); })
      .map(function(e) { return e.target.value; })
      .filter(function(text) { return text.length > 2; })
      .distinctUntilChanged()
      .throttle(1000);

    // Merge observables
    var observable = Rx.Observable.merge(clickSource, inputSource);

    // Iterate providers and create observable for each provider
    providerList.forEach(function(provider) {
      observable
        .flatMapLatest(function(keyword) {
          searchPrepare();
          return Rx.Observable.onErrorResumeNext(Rx.Observable.fromPromise(
            search(provider.name, keyword).then(
              function(data) { return data; },
              function(err) { searchError(err, provider); }
            )
          ));
        })
        .subscribe(
          function(data) {
            searchResults(data, provider);
          },
          function(err) {
            searchError(err, provider);
          }
        );
    });
  }

  // providersGet gets the provider list
  function providersGet() {
    return $.ajax({
      url:      serverUrl+appPath+'providers',
      dataType: 'jsonp',
      method:   'GET',
    }).promise();
  }

  // providersPrepare prepares UI for providers
  function providersPrepare() {
    // Iterate providers and prepare DOM elements
    providerList.forEach(function(provider) {
      // Navigations
      $('#searchResultsNavs').append('<li id="searchResultsNavs-' + provider.name + '" class="disabled" role="presentation"><a href="#' + provider.name + '" aria-controls="' + provider.name + '" role="tab" data-toggle="tab">' + provider.title + '</a></li>');
      // Tabs
      $('#searchResultsTabs').append('<div id="' + provider.name + '" role="tabpanel" class="tab-pane">');
    });

    // Initialize navigation clicks
    $('#searchResultsNavs a').click(function (e) {
      e.preventDefault();
      $(this).tab('show');
    });
  }

  // search makes a search by the given provider and keyword
  function search(provider, keyword) {
    return $.ajax({
      url:      serverUrl+appPath+'search',
      dataType: 'jsonp',
      method:   'GET',
      data: {
        provider: (''+provider),
        keyword:  (''+keyword),
        timeout:  '5000ms',
        limit:    10
      }
    }).promise();
  }

  // searchReset resets UI for search
  function searchReset() {
    providerList.forEach(function(provider) { $('#' + provider.name).empty(); });

    $('#searchInput').val('');
    $('#searchResultsNavs').addClass('search-result-navs-hide');
    $('#logoMain').detach().appendTo($('#logoMiddleHolder')).removeClass('logo-navbar');
    $('#searchMain').detach().appendTo($("#searchMiddleHolder"));
    $('#searchInput').focus();
  }

  // searchPrepare prepares UI for search
  function searchPrepare() {
    providerList.forEach(function(provider) { $('#' + provider.name).empty(); });

    $('#logoMain').detach().appendTo($('#logoNavbarHolder')).addClass('logo-navbar');
    $('#searchMain').detach().appendTo($("#searchNavbarHolder"));
    $('#searchResultsNavs').removeClass('search-result-navs-hide');

    // If the current search input value is different than previous value then
    var sivc = $('#searchInput').val();
    if(sivc !== $('#searchInput').attr('data-prev-value')) {
      if($('#searchInput').attr('data-caller') !== 'onpopstate') {
        window.history.pushState({q: sivc}, null, appPath+'?q=' + sivc + location.hash);
      }
      $('#searchInput').attr('data-caller', null);
      $(document).prop('title', 'Ferret - ' + sivc);
    }
    $('#searchInput').attr('data-prev-value', sivc);
    $('#searchInput').focus();
  }

  // searchResults renders search results
  function searchResults(data, provider) {
    if(data && data instanceof Array) {
      // Prepare result content
      var content = '';
      if(provider && typeof provider === 'object') {
        // Iterate results
        $.map(data, function (v) {
          content += '<li class="search-results-li">';
          content += '<a href="' + v.link + '" target="_blank">' + v.title + '</a>';
          content += '<p>';
          content += (v.description) ? encodeHtmlEntity(v.description)+'<br>' : '';
          content += (v.date != "0001-01-01T00:00:00Z") ? '<span class="ts">' + (''+(new Date(v.date)).toISOString()).substr(0, 10) + '</span>' : '';
          content += '</p>';
          content += '</li>';
        });
      }
      $('#' + provider.name).html(content);
      $('#' + provider.name).attr('data-result', true);
      $('#searchResultsNavs-' + provider.name).removeClass("disabled");
    } else {
      $('#' + provider.name).html('No results');
      $('#' + provider.name).attr('data-result', false);
      $('#searchResultsNavs-' + provider.name).addClass("disabled");
    }

    var tId = $('#searchResultsTabs [data-result=true]:first').attr("id");
    if(tId) {
      $('a[role=tab][aria-controls=' + tId + ']').tab('show');
    }
  }

  // searchError shows a search error
  function searchError(err, provider) {
    var e = parseError(err);
    if(provider && typeof provider === 'object') {
      $('#' + provider.name).append($('<div class="alert alert-danger" role="alert">').text(e.message));
    } else {
      $('#searchResultsTabs').append($('<div class="alert alert-danger" role="alert">').text(e.message));
    }
  }

  // warning shows a warning message
  // function warning(message) {
  //   $('#searchAlerts').html($('<div class="alert alert-warning search-alert" role="alert">').text(message));
  // }

  // critical shows a critical message
  function critical(message) {
    $('#searchAlerts').html($('<div class="alert alert-danger search-alert" role="alert">').text(message));
  }

  // parseError parses the given error message and returns an object
  function parseError(err) {
    var code    = 0,
        message = 'unknown error';

    if(err && typeof err == 'object') {
      code    = err.status || code;
      message = err.statusText || err.message || message;
      if(typeof err.responseJSON == 'object') {
        message = err.responseJSON.message || err.responseJSON.error || message;
      }
    }

    return {code: code, message: message};
  }

  // encodeHtmlEntity encodes HTML entity
  function encodeHtmlEntity(str) {
    return str.replace(/[\u00A0-\u9999\\<\>\&\'\"\\\/]/gim, function(c) {
      return '&#' + c.charCodeAt(0) + ';';
    });
  }

  // urlParam returns URL query parameter by the given name
  function urlParam(name, url) {
    url = url || location.href;
    var results = new RegExp('[\?&]' + name + '=([^&#]*)').exec(url);
    if(!results) {
      return;
    }
    return results[1] || null;
  }

  // urlQueryHandler handles url query parameter
  function urlQueryHandler() {
    // Handle back/forward buttons
    window.onpopstate = function(/*event*/) {
      // If query parameter is not empty then
      if(urlParam('q')) {
        // Set input value and trigger click
        $('#searchInput').attr('data-caller', 'onpopstate');
        $('#searchInput').val(unescape(urlParam('q')));
        $('#searchButton').click();
      } else {
        searchReset();
      }
    };

    // If query parameter is not empty then
    if(urlParam('q')) {
      // Set input value and trigger click
      $('#searchInput').val(unescape(urlParam('q')));
      $('#searchButton').click();
    }

    $('#searchInput').focus();
  }

  // Return
  return {
    init: init
  };
};

$(document).ready(function() {
  app().init();
});
