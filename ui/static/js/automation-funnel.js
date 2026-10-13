(() => {
  const campaignKeys = ["utm_source", "utm_medium", "utm_campaign", "utm_content"];
  const query = new URLSearchParams(window.location.search);
  const campaign = Object.fromEntries(
    campaignKeys.map((key) => [key, query.get(key) || ""]),
  );

  const track = (event, details = {}) => {
    window.dataLayer = window.dataLayer || [];
    window.dataLayer.push({ event, ...details });
  };

  const assessmentForm = document.querySelector("#automation-assessment-form");
  if (assessmentForm) {
    const assessment = document.querySelector("#assessment-result");
    const result = document.querySelector("#post-assessment-result");
    const contactCTA = document.querySelector("#assessment-contact-cta");
    let started = false;

    track("automation_assessment_view", campaign);

    assessmentForm.addEventListener("input", () => {
      if (started) {
        return;
      }
      started = true;
      track("automation_assessment_start", campaign);
    });

    assessmentForm.addEventListener("submit", (event) => {
      event.preventDefault();

      const data = new FormData(assessmentForm);
      const fields = ["frequency", "consistency", "systems", "impact", "oversight"];
      const answers = Object.fromEntries(fields.map((field) => {
        const input = assessmentForm.querySelector(`input[name="${field}"]:checked`);
        return [field, {
          score: Number(data.get(field)),
          label: input.closest("label").querySelector("span").textContent.trim(),
        }];
      }));
      const score = fields.reduce((total, field) => total + answers[field].score, 0);
      const band = readinessBand(score, answers.oversight.score);
      const workflowName = data.get("workflow_name").trim();

      document.querySelector("#assessment-result-icon").textContent = band.icon;
      document.querySelector("#assessment-result-label").textContent = `${band.label} · ${score}/10`;
      document.querySelector("#assessment-result-heading").textContent = band.heading;
      document.querySelector("#assessment-result-heading-emphasis").textContent = band.headingEmphasis;
      document.querySelector("#assessment-result-copy").textContent = band.copy;

      const factorList = document.querySelector("#assessment-result-factors");
      factorList.replaceChildren(...buildFactors(answers).map((factor) => {
        const item = document.createElement("li");
        item.textContent = factor;
        return item;
      }));

      const contactQuery = new URLSearchParams({
        workflow_name: workflowName,
        assessment_result: band.key,
        assessment_score: String(score),
        assessment_answers: JSON.stringify(answers),
      });
      campaignKeys.forEach((key) => {
        if (campaign[key]) {
          contactQuery.set(key, campaign[key]);
        }
      });
      if (document.referrer) {
        contactQuery.set("referrer", new URL(document.referrer).origin);
      }
      contactCTA.href = `/contact?${contactQuery.toString()}`;
      contactCTA.textContent = band.cta;
      contactCTA.dataset.assessmentResult = band.key;

      assessment.classList.add("hidden");
      result.classList.remove("hidden");
      result.focus();
      result.scrollIntoView({
        behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "auto" : "smooth",
        block: "center",
      });
      track("automation_assessment_complete", {
        ...campaign,
        assessment_result: band.key,
        assessment_score: score,
      });
    });

    contactCTA.addEventListener("click", () => {
      track("automation_assessment_contact_click", {
        ...campaign,
        assessment_result: contactCTA.dataset.assessmentResult,
      });
    });
  }

  const contactForm = document.querySelector("#contact-form");
  if (contactForm) {
    contactForm.addEventListener("submit", () => {
      track("commercial_contact_submit", {
        ...campaign,
        lead_source: contactForm.dataset.leadSource || "direct",
        assessment_result: contactForm.elements.assessment_result.value,
      });
    });
  }

  const contactSuccess = document.querySelector("#contact-success");
  if (contactSuccess) {
    track("commercial_contact_success", {
      lead_source: contactSuccess.dataset.leadSource || "direct",
    });
  }

  function readinessBand(score, oversight) {
    if (score >= 8 && oversight > 0) {
      return {
        key: "strong",
        icon: "rocket_launch",
        label: "Strong Automation Candidate",
        heading: "This workflow is",
        headingEmphasis: "ready for a closer look.",
        copy: "Your responses show the repetition, structure, digital access, business impact, and human oversight that support secure automation.",
        cta: "Discuss This Workflow",
      };
    }
    if (score >= 4) {
      return {
        key: "emerging",
        icon: "trending_up",
        label: "Emerging Automation Candidate",
        heading: "This workflow has",
        headingEmphasis: "potential with some preparation.",
        copy: "Parts of this workflow are ready to automate, while the weaker areas should be clarified before implementation.",
        cta: "Review This Workflow With Us",
      };
    }
    return {
      key: "foundation",
      icon: "handyman",
      label: "Build the Foundation First",
      heading: "Standardize this workflow",
      headingEmphasis: "before automating it.",
      copy: "Automation will be more reliable after the process, inputs, ownership, and expected outcomes become more consistent.",
      cta: "Get a Practical Readiness Plan",
    };
  }

  function buildFactors(answers) {
    const factors = [];
    factors.push(answers.frequency.score >= 1
      ? `Volume: ${answers.frequency.label} creates a meaningful opportunity to return time to the team.`
      : `Volume: ${answers.frequency.label} may limit the immediate return on automation.`);
    factors.push(answers.consistency.score >= 1
      ? `Process: ${answers.consistency.label} provides a basis for defining reliable automation rules.`
      : `Process: ${answers.consistency.label} suggests documenting a repeatable core before implementation.`);
    factors.push(answers.systems.score >= 1
      ? `Systems: ${answers.systems.label} makes controlled digital integration more feasible.`
      : `Systems: ${answers.systems.label} means the inputs should be digitized or structured first.`);
    factors.push(answers.impact.score >= 1
      ? `Business case: ${answers.impact.label} indicates a measurable reason to improve the workflow.`
      : `Business case: ${answers.impact.label} suggests prioritizing a higher-impact workflow.`);
    factors.push(answers.oversight.score >= 1
      ? `Governance: ${answers.oversight.label} supports human review of exceptions and sensitive decisions.`
      : `Governance: ${answers.oversight.label} must be resolved before sensitive work is automated.`);
    return factors;
  }

})();
