import type { Modal } from "bootstrap";

type BSModalOptions = Partial<Modal.Options> & {
  onHide?: () => void;
  onShow?: () => void;
  onShown?: () => void;
  onHidden?: () => void;
};

export function useModal(modalEl: HTMLElement, options: BSModalOptions = {}) {
  // Allow Esc to close and let Bootstrap move focus.
  const modalOptions: Partial<Modal.Options> = { backdrop: "static", keyboard: true, focus: true, ...options };
  const modalInstance = $(modalEl).modal({
    backdrop: modalOptions.backdrop,
    keyboard: modalOptions.keyboard,
    focus: modalOptions.focus,
    // @ts-ignore silence false error due to wrong typping
    show: false
  })

  // Remember the element that had focus before opening, to restore it on close.
  let lastTrigger: HTMLElement | null = null;

  if (options.onShow) {
    $(modalEl).on("show.bs.modal", function () {
      options.onShow && options.onShow();
    });
  }
  $(modalEl).on("shown.bs.modal", function () {
    // Move focus into the modal so Esc works and screen readers enter the dialog
    const focusTarget =
      modalEl.querySelector<HTMLElement>('[autofocus], [data-initial-focus], input, button, select, textarea, [tabindex]:not([tabindex="-1"])')
      || modalEl;
    focusTarget.focus();
    options.onShown && options.onShown();
  });

  $(modalEl).on("hide.bs.modal", function () {
    const active = document.activeElement as HTMLElement | null
    if (active && modalEl?.contains(active)) {
      active.blur() // or triggerButton.focus()
    }
    options.onHide && options.onHide();
  });
  if (options.onHidden) {
    $(modalEl).on("hidden.bs.modal", function () {
      options.onHidden && options.onHidden();
      hideLoading();
      // Restore focus to the opener after fully hidden
      lastTrigger?.focus();
      lastTrigger = null;
    });
  }

  const headerEl = modalEl?.querySelector(".modal-header") as HTMLElement | null | undefined;
  const titleEl = headerEl?.querySelector(".modal-title") as HTMLElement | null | undefined;

  // 
  let loadingOverlay: HTMLElement | null = null;

  function showLoading() {
    loadingOverlay = createLoadingOverlay();
    loadingOverlay.style.top = headerEl ? `${headerEl.offsetHeight}px` : '0';
    modalEl.querySelector('.modal-content')?.appendChild(loadingOverlay);
  }

  function hideLoading() {
    if (loadingOverlay) {
      loadingOverlay.remove();
      loadingOverlay = null;
    }
  }

  return {
    setOptions(config?: Partial<Modal.Options>) {
      const instance = modalInstance?.data("bs.modal");
      const instanceConfig = modalInstance?.data("bs.modal")?._config as Modal.Options | undefined;
      instance && (instance._config = { ...instanceConfig, ...config });
    },
    setTitle(title: string) {
      titleEl && (titleEl.textContent = title);
      headerEl?.classList.remove("d-none")
    },
    hideHeader() {
      headerEl?.classList.add("d-none");
    },
    // Prefer passing the actual trigger element. Falls back to activeElement.
    open(trigger?: HTMLElement) {
      lastTrigger = trigger ?? (document.activeElement as HTMLElement | null);
      modalInstance && modalInstance.modal("show");
    },
    close() {
      modalInstance && modalInstance.modal("hide");
    },
    showLoading,
    hideLoading
  };
}

/** Helper function to create loading overlay */
function createLoadingOverlay(): HTMLElement {
  const overlay = document.createElement('div');
  overlay.style.cssText = `
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(255, 255, 255, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1050;
    backdrop-filter: blur(1px);
  `;

  const spinner = document.createElement('div');
  spinner.innerHTML = `
    <div class="d-flex flex-column align-items-center">
      <div class="spinner-border text-secondary" role="status" style="width: 3rem; height: 3rem;">
        <span class="sr-only">Loading...</span>
      </div>
      <div class="mt-2 text-muted small">Processing...</div>
    </div>
  `;

  overlay.appendChild(spinner);
  return overlay;
}