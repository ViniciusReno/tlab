(() => {
  const form = document.getElementById("scenario-form");
  const yieldInput = document.getElementById("yield");
  const slider = document.getElementById("yield-slider");
  let timer;
  let activeRequest;
  let generation = 0;

  slider.addEventListener("input", () => {
    yieldInput.value = slider.value;
    clearTimeout(timer);
    timer = setTimeout(() => form.requestSubmit(), 180);
  });
  yieldInput.addEventListener("input", () => { slider.value = yieldInput.value; });
  document.addEventListener("click", (event) => {
    const shock = event.target.closest("[data-yield]");
    if (!shock) return;
    event.preventDefault();
    yieldInput.value = shock.dataset.yield;
    slider.value = yieldInput.value;
    form.requestSubmit();
  });
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    clearTimeout(timer);
    if (activeRequest) activeRequest.abort();
    activeRequest = new AbortController();
    const current = ++generation;
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
      results.replaceWith(replacement);
      if (wasOpen && document.getElementById("technical")) document.getElementById("technical").open = true;
      history.replaceState(null, "", target);
    } catch (error) {
      if (error.name !== "AbortError" && current === generation) window.location.assign(target);
    } finally {
      if (current === generation) document.getElementById("results").removeAttribute("aria-busy");
    }
  });
})();
