(() => {
  let isLoaded = false;

  const loadAnalytics = () => {
    if (isLoaded) {
      return;
    }

    isLoaded = true;
    window.dataLayer = window.dataLayer || [];
    window.dataLayer.push({ "gtm.start": Date.now(), event: "gtm.js" });

    const script = document.createElement("script");
    script.async = true;
    script.src = "https://www.googletagmanager.com/gtm.js?id=GTM-WH3CMBPJ";
    document.head.appendChild(script);
  };

  const interactionEvents = ["keydown", "pointerdown", "touchstart"];
  interactionEvents.forEach((eventName) => {
    window.addEventListener(eventName, loadAnalytics, { once: true, passive: true });
  });

  window.addEventListener("load", () => {
    window.setTimeout(loadAnalytics, 3500);
  }, { once: true });
})();