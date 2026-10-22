(() => {
  const consentKey = "tabiri-analytics-consent";
  const consentBanner = document.getElementById("analytics-consent");
  const acceptButton = document.getElementById("accept-analytics");
  const rejectButton = document.getElementById("reject-analytics");
  const settingsButton = document.getElementById("cookie-settings");
  let analyticsLoaded = false;
  let consent = null;

  const updateGoogleConsent = (value) => {
    if (!window.dataLayer) {
      return;
    }

    window.gtag = window.gtag || function () {
      window.dataLayer.push(arguments);
    };
    window.gtag("consent", "update", {
      analytics_storage: value,
      ad_storage: "denied",
      ad_user_data: "denied",
      ad_personalization: "denied",
    });
  };

  const loadAnalytics = () => {
    if (analyticsLoaded || consent !== "accepted") {
      return;
    }

    analyticsLoaded = true;
    window.dataLayer = window.dataLayer || [];
    window.gtag = window.gtag || function () {
      window.dataLayer.push(arguments);
    };
    window.gtag("consent", "default", {
      analytics_storage: "denied",
      ad_storage: "denied",
      ad_user_data: "denied",
      ad_personalization: "denied",
    });
    updateGoogleConsent("granted");
    window.dataLayer.push({ "gtm.start": Date.now(), event: "gtm.js" });

    const script = document.createElement("script");
    script.async = true;
    script.src = "https://www.googletagmanager.com/gtm.js?id=GTM-WH3CMBPJ";
    document.head.appendChild(script);
  };

  const showConsentBanner = () => {
    consentBanner.hidden = false;
  };

  const saveConsent = (value) => {
    consent = value;
    try {
      window.localStorage.setItem(consentKey, value);
    } catch (error) {
      console.error("Could not save analytics consent preference", error);
    }
    consentBanner.hidden = true;

    if (value === "accepted") {
      loadAnalytics();
    } else if (analyticsLoaded) {
      updateGoogleConsent("denied");
    }
  };

  try {
    const savedConsent = window.localStorage.getItem(consentKey);
    if (savedConsent === "accepted" || savedConsent === "rejected") {
      consent = savedConsent;
    }
  } catch (error) {
    console.error("Could not read analytics consent preference", error);
  }

  if (consent === "accepted") {
    loadAnalytics();
  } else if (consent === null) {
    showConsentBanner();
  }

  acceptButton.addEventListener("click", () => saveConsent("accepted"));
  rejectButton.addEventListener("click", () => saveConsent("rejected"));
  settingsButton.addEventListener("click", () => {
    showConsentBanner();
    rejectButton.focus();
  });
})();
