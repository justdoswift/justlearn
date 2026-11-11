(() => {
  const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  if (reducedMotion || !window.gsap || !window.ScrollTrigger) return;

  gsap.registerPlugin(ScrollTrigger);

  gsap.from(".title-line > span", {
    yPercent: 110,
    duration: 1.05,
    ease: "expo.out",
    stagger: 0.12,
  });

  gsap.from(".hero-copy, .hero-foot", {
    y: 16,
    duration: 0.8,
    ease: "power3.out",
    stagger: 0.1,
    delay: 0.42,
  });

  gsap.from(".hero-mark", {
    opacity: 0,
    scale: 0.92,
    duration: 1.3,
    ease: "power2.out",
    delay: 0.2,
  });

  gsap.fromTo(
    ".route-progress",
    { scaleY: 0 },
    {
      scaleY: 1,
      transformOrigin: "top center",
      ease: "none",
      scrollTrigger: {
        trigger: ".route-track",
        start: "top 72%",
        end: "bottom 48%",
        scrub: true,
      },
    }
  );

  gsap.from(".route-stage", {
    y: 16,
    duration: 0.75,
    ease: "power3.out",
    stagger: 0.1,
    scrollTrigger: {
      trigger: ".route-track",
      start: "top 76%",
      once: true,
    },
  });

  gsap.from(".route-diagram .diagram-note", {
    y: 10,
    duration: 0.6,
    ease: "power2.out",
    scrollTrigger: {
      trigger: ".route-diagram",
      start: "top 72%",
      once: true,
    },
  });

  gsap.utils.toArray("[data-reveal]").forEach((element) => {
    gsap.from(element, {
      y: 24,
      duration: 0.8,
      ease: "power3.out",
      scrollTrigger: {
        trigger: element,
        start: "top 84%",
        once: true,
      },
    });
  });

  gsap.from(".rhythm-list li", {
    y: 20,
    duration: 0.8,
    ease: "power3.out",
    stagger: 0.09,
    scrollTrigger: {
      trigger: ".rhythm-list",
      start: "top 80%",
      once: true,
    },
  });

  gsap.from(".ability-list li", {
    y: 16,
    duration: 0.7,
    ease: "power3.out",
    stagger: 0.07,
    scrollTrigger: {
      trigger: ".ability-board",
      start: "top 84%",
      once: true,
    },
  });
})();
