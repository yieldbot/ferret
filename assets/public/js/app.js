/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

/* jslint browser: true */
/* global document: false, $: false, Rx: false */
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
          // Sort the list
          providerList = data.sort(function(a, b) {
            return b.priority - a.priority;
          });
        }
        // Listen for search requests
        listen(providerList);
      },
      function(err) {
        var e = parseError(err);
        critical("Could not get the available providers due to " + e.message  + " (" + e.code + ")");
      }
    );
  }

  // listen listens search requests for the given providers
  function listen(providers) {
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

    // Check providers
    if(!(providers instanceof Array) || providers.length < 1) {
      critical("There is no any available provider to search");
      return;
    }
    providersPrepare();

    // Iterate providers
    providers.forEach(function(provider) {
      observable
        .flatMapLatest(function(keyword) {
          searchPrepare();

          // Exceptions
          keyword = (provider.name == "github") ? keyword+'+extension:md' : keyword;

          // Search
          var searchSource = Rx.Observable.fromPromise(
            search(provider.name, keyword).then(
              function(data) {
                return data;
              },
              function(err) {
                searchError(err, provider);
              }
            )
          );

          return Rx.Observable.onErrorResumeNext(searchSource);
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

    $('#searchInput').focus();
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

    // Iterate providers
    providerList.forEach(function(provider) {
      // Prepare navigation bar tabs
      var content = '<li role="presentation">';
      content += '<a href="javascript:void(0)" onclick="$.scrollTo(\'#' + provider.name + '\', 500, {offset: {top: -110}})">' + provider.title + '</a>';
      content += '</li>';
      $('#searchNavbarTabs').append(content);

      // Prepare search result divs
      $('#searchResults').append('<div id="searchResults-' + provider.name + '">');
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

  // searchPrepare prepares UI for search
  function searchPrepare() {
    $('#logoMain').detach().appendTo($('#logoNavbarHolder')).addClass('logo-navbar');
    $('#searchMain').detach().appendTo($("#searchNavbarHolder")).addClass('input-group-search-navbar');
    $('#searchNavbar').removeClass('search-navbar-hide');
    $('div[id^=searchResults-]').empty();
  }

  // searchResults renders search results
  function searchResults(data, provider) {
    if(data && data instanceof Array) {

      // Prepare result content
      var content = '';
      if(provider && typeof provider === 'object') {
        content += '<h3 id="' + provider.name + '">' + provider.title + '</h3>';

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
        content += '<hr>';
      }
      $("#searchResults-" + provider.name).html(content);
    }
  }

  // searchError shows a search error
  function searchError(err, provider) {
    var e = parseError(err);
    if(provider && typeof provider === 'object') {
      $('#searchResults').append($('<h3>').text(provider.title));
    }
    $('#searchResults').append($('<div class="alert alert-danger" role="alert">').text(e.message));
  }

  // warning shows a warning message
  function warning(message) {
    $('#searchAlerts').html($('<div class="alert alert-warning search-alert" role="alert">').text(message));
  }

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
    return str.replace(/[\u00A0-\u9999\\<\>\&\'\"\\\/]/gim, function(c){
      return '&#' + c.charCodeAt(0) + ';' ;
    });
  }

  // Return
  return {
    init: init
  };
};

$(document).ready(function() {
  app().init();
});
