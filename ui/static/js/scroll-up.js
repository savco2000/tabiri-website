(() => {
  const scrollBtn = document.getElementById('scrollUpBtn');

  if (!scrollBtn) {
    return;
  }

  scrollBtn.addEventListener('click', () => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  });

  window.addEventListener('scroll', () => {
    if (window.scrollY > 300) {
      scrollBtn.hidden = false;
      scrollBtn.classList.add('scroll-up--visible', 'scroll-up--raised');
    } else {
      scrollBtn.classList.remove('scroll-up--raised');
      setTimeout(() => {
        if (window.scrollY <= 300) scrollBtn.hidden = true;
      }, 300);
    }
  });
})();
