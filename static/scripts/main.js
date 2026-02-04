// Statigo - Minimal JavaScript
// This file is intentionally minimal. Add your application logic here.

(function() {
  'use strict';

  // DOM Ready helper
  function ready(fn) {
    if (document.readyState !== 'loading') {
      fn();
    } else {
      document.addEventListener('DOMContentLoaded', fn);
    }
  }

  // Initialize on DOM ready
  ready(function() {
    console.log('Statigo initialized');
  });
})();
