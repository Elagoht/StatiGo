// Statigo - Minimal JavaScript

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
    const counterBtn = document.getElementById('counterBtn');
    const counterValue = document.getElementById('counterValue');

    if (counterBtn && counterValue) {
      let count = 0;
      counterBtn.addEventListener('click', function() {
        count++;
        counterValue.textContent = count;
      });
    }
  });
})();
