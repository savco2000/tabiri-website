(() => {
  const triggerSelector = '.nav-dropdown-trigger';
  const menuSelector = '.nav-dropdown-menu';

  const initDropdownNav = () => {
    const triggers = Array.from(document.querySelectorAll(triggerSelector));
    const menus = Array.from(document.querySelectorAll(menuSelector));

    if (triggers.length === 0 || menus.length === 0) {
      return;
    }

    const closeMenus = () => {
      menus.forEach((menu) => {
        menu.classList.add('hidden');
        menu.setAttribute('aria-hidden', 'true');
      });

      triggers.forEach((trigger) => {
        trigger.setAttribute('aria-expanded', 'false');
      });
    };

    const setActiveDropdown = () => {
      const currentPath = window.location.pathname;

      triggers.forEach((trigger) => {
        const targetId = trigger.getAttribute('data-dropdown-target');
        const targetMenu = targetId ? document.getElementById(targetId) : null;

        if (!targetMenu) {
          return;
        }

        const isActive = Array.from(targetMenu.querySelectorAll('a')).some((link) => {
          const href = link.getAttribute('href');
          if (!href) {
            return false;
          }

          const linkPath = new URL(href, window.location.href).pathname;
          return link.classList.contains('nav-link-active') || linkPath === currentPath;
        });

        trigger.classList.toggle('nav-dropdown-trigger-active', isActive);
      });
    };

    setActiveDropdown();
    closeMenus();

    triggers.forEach((trigger) => {
      const targetId = trigger.getAttribute('data-dropdown-target');
      const targetMenu = targetId ? document.getElementById(targetId) : null;

      if (!targetMenu) {
        return;
      }

      trigger.setAttribute('aria-controls', targetId);
      trigger.setAttribute('aria-expanded', 'false');
      targetMenu.setAttribute('aria-hidden', 'true');
    });

    document.body.addEventListener('click', (event) => {
      const target = event.target;

      if (!(target instanceof Element)) {
        return;
      }

      const trigger = target.closest(triggerSelector);
      const menu = target.closest(menuSelector);

      if (trigger) {
        event.stopPropagation();
        const targetId = trigger.getAttribute('data-dropdown-target');
        const targetMenu = targetId ? document.getElementById(targetId) : null;

        if (!targetMenu) {
          return;
        }

        const isOpen = !targetMenu.classList.contains('hidden');
        closeMenus();

        if (!isOpen) {
          targetMenu.classList.remove('hidden');
          targetMenu.setAttribute('aria-hidden', 'false');
          trigger.setAttribute('aria-expanded', 'true');
        }
        return;
      }

      if (!menu) {
        closeMenus();
      }
    });

    document.addEventListener('keydown', (event) => {
      if (event.key === 'Escape') {
        closeMenus();
      }
    });
  };

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initDropdownNav, { once: true });
  } else {
    initDropdownNav();
  }
})();
