(() => {
  const scrollBtn = document.getElementById('scrollUpBtn');
  window.addEventListener('scroll', () => {
    if (window.scrollY > 300) {
      scrollBtn.classList.remove('hidden', 'translate-y-4', 'opacity-0');
      scrollBtn.classList.add('translate-y-0', 'opacity-100');
    } else {
      scrollBtn.classList.add('translate-y-4', 'opacity-0');
      setTimeout(() => {
        if (window.scrollY <= 300) scrollBtn.classList.add('hidden');
      }, 300);
    }
  });
})();
