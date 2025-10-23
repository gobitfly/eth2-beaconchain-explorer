import type { Modal } from "bootstrap";

type BSModalOptions = Partial<Modal.Options> & {
  onHide?: () => void;
  onShow?: () => void;
  onSHown?: () => void;
  onHidden?: () => void;
};

export function useBSModal(modalEl: HTMLElement, options: BSModalOptions = {}) {
  const modalOptions: Partial<Modal.Options> = { backdrop: "static", keyboard: false, focus: false, ...options };
  const modalInstance = $(modalEl).modal({
    backdrop: modalOptions.backdrop,
    keyboard: modalOptions.keyboard,
    focus: modalOptions.focus,
    // @ts-ignore silence false error due to wrong typping
    show: false
  })

  if (options.onShow) {
    $(modalEl).on("show.bs.modal", function () {
      options.onShow && options.onShow();
    });
  }
  if (options.onSHown) {
    $(modalEl).on("shown.bs.modal", function () {
      options.onSHown && options.onSHown();
    });
  }
  if (options.onHide) {
    $(modalEl).on("hide.bs.modal", function () {
      options.onHide && options.onHide();
    });
  }
  if (options.onHidden) {
    $(modalEl).on("hidden.bs.modal", function () {
      options.onHidden && options.onHidden();
      hideLoading();
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
    setOptiopns(config?: Partial<Modal.Options>) {
      const instance = modalInstance?.data("bs.modal");
      const instanceConfig = modalInstance?.data("bs.modal")?._config as Modal.Options | undefined;
      instance && (instance._config = { ...instanceConfig, ...config });
    },
    seTitle(title: string) {
      titleEl && (titleEl.textContent = title);
      headerEl?.classList.remove("d-none")
    },
    hideHeader() {
      headerEl?.classList.add("d-none");
    },
    open() {
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