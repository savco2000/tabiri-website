(() => {
  const mobileNavToggle = document.getElementById("mobile-nav-toggle");
  const mobileNavPanel = document.getElementById("mobile-nav-panel");

  if (!mobileNavToggle || !mobileNavPanel) {
    return;
  }

  const mainContent = document.querySelector("main");
  const mobileNavLinks = mobileNavPanel.querySelectorAll("a");

  const getFocusableInPanel = () =>
    Array.from(
      mobileNavPanel.querySelectorAll('a, button, [tabindex]:not([tabindex="-1"])')
    );

  const isMobileNavOpen = () =>
    mobileNavToggle.getAttribute("aria-expanded") === "true";

  const closeMobileNav = () => {
    mobileNavPanel.classList.remove("is-open");
    mobileNavPanel.setAttribute("aria-hidden", "true");
    mobileNavToggle.setAttribute("aria-expanded", "false");
    mobileNavToggle.setAttribute("aria-label", "Open site navigation");
    mobileNavToggle.querySelector(".material-symbols-outlined").textContent = "menu";
    document.body.classList.remove("overflow-hidden");

    if (mainContent) {
      mainContent.removeAttribute("inert");
    }
  };

  const openMobileNav = () => {
    mobileNavPanel.classList.add("is-open");
    mobileNavPanel.setAttribute("aria-hidden", "false");
    mobileNavToggle.setAttribute("aria-expanded", "true");
    mobileNavToggle.setAttribute("aria-label", "Close site navigation");
    mobileNavToggle.querySelector(".material-symbols-outlined").textContent = "close";
    document.body.classList.add("overflow-hidden");

    if (mainContent) {
      mainContent.setAttribute("inert", "");
    }

    const focusable = getFocusableInPanel();
    if (focusable.length > 0) {
      focusable[0].focus();
    }
  };

  mobileNavToggle.addEventListener("click", () => {
    if (isMobileNavOpen()) {
      closeMobileNav();
    } else {
      openMobileNav();
    }
  });

  mobileNavLinks.forEach((link) => {
    link.addEventListener("click", closeMobileNav);
  });

  document.addEventListener("keydown", (event) => {
    if (!isMobileNavOpen()) {
      return;
    }

    if (event.key === "Escape") {
      closeMobileNav();
      mobileNavToggle.focus();
      return;
    }

    if (event.key === "Tab") {
      const focusable = getFocusableInPanel();
      if (focusable.length === 0) {
        event.preventDefault();
        return;
      }

      const first = focusable[0];
      const last = focusable[focusable.length - 1];

      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }
  });

  document.addEventListener("click", (event) => {
    if (!isMobileNavOpen()) {
      return;
    }

    if (!mobileNavPanel.contains(event.target) && !mobileNavToggle.contains(event.target)) {
      closeMobileNav();
    }
  });

  window.addEventListener("resize", () => {
    if (window.innerWidth >= 1024) {
      closeMobileNav();
    }
  });
})();
