(() => {
  const form = document.getElementById("scenario-form");
  const yieldInput = document.getElementById("yield");
  const slider = document.getElementById("yield-slider");
  if (!form || !yieldInput) return;
  if (slider) document.getElementById("slider-control").hidden = false;
  const contextFields = ["bond", "basis", "date"];
  const initialContext = contextFields.map(name => form.elements.namedItem(name).value);
  let sliderSubmission = false;
  let timer;
  let activeRequest;
  let generation = 0;

  slider?.addEventListener("input", () => {
    yieldInput.value = slider.value;
    clearTimeout(timer);
    timer = setTimeout(() => { sliderSubmission = true; form.requestSubmit(); sliderSubmission = false; }, 180);
  });
  yieldInput.addEventListener("input", () => { if (slider) slider.value = yieldInput.value; });
  document.addEventListener("click", (event) => {
    const shock = event.target.closest("[data-yield]");
    if (!shock) return;
    event.preventDefault();
    yieldInput.value = shock.dataset.yield;
    if (slider) slider.value = yieldInput.value;
    form.requestSubmit();
  });
  form.addEventListener("submit", async (event) => {
    // A new quote context also changes labels and the slider's official baseline.
    // Let the server render the complete page for that change.
    if (contextFields.some((name, index) => form.elements.namedItem(name).value !== initialContext[index])) return;
    event.preventDefault();
    clearTimeout(timer);
    if (activeRequest) activeRequest.abort();
    activeRequest = new AbortController();
    const current = ++generation;
    const moveFocus = !sliderSubmission;
    const query = new URLSearchParams(new FormData(form));
    const target = form.action + "?" + query.toString();
    const results = document.getElementById("results");
    results.setAttribute("aria-busy", "true");
    try {
      const response = await fetch(target, { signal: activeRequest.signal });
      const html = new DOMParser().parseFromString(await response.text(), "text/html");
      const replacement = html.getElementById("results");
      if (current !== generation) return;
      if (!replacement) throw new Error("Missing result");
      const wasOpen = document.getElementById("technical")?.open;
      results.replaceChildren(...replacement.childNodes);
      if (moveFocus) results.focus({ preventScroll: true });
      if (wasOpen && document.getElementById("technical")) document.getElementById("technical").open = true;
      history.replaceState(null, "", target);
      // A language change must retain the scenario most recently calculated.
      document.querySelectorAll(".language-picker a").forEach(link => {
        const next = new URL(target, window.location.href);
        next.searchParams.set("lang", link.hreflang);
        link.href = next.pathname + next.search;
      });
    } catch (error) {
      if (error.name !== "AbortError" && current === generation) window.location.assign(target);
    } finally {
      if (current === generation) document.getElementById("results").removeAttribute("aria-busy");
    }
  });
})();
