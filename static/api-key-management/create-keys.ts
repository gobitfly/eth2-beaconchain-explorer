import type { ApiKeysClient, APIKey, CreateAPIKeyData } from "./api";
import { useModal } from "./modal";
import { copyTextWithToast, validateApiKeyName, renderKeyData, clearKeyData } from "./utils";

const DEFAULTOPTIONS = {
  // Form options
  createFormSelector: "#createKeyForm",
  keyNameinputSelector: "#newApiKeyField",
  submitBtnSelector: "[type=submit]",
  rulesSelector: 'ul.list-unstyled [data-validate-rule]',
};

export function useAPIKeyCreation(api: ApiKeysClient, options: Partial<typeof DEFAULTOPTIONS> & { onCreated?: () => void; } = {}) {
  const config = { ...DEFAULTOPTIONS, ...options };

  const createForm = document.querySelector(config.createFormSelector) as HTMLFormElement | null;
  const keyNameinput = document.querySelector(config.keyNameinputSelector) as HTMLInputElement | null;

  if (!createForm || !keyNameinput) return;

  const validateNameAndSetStatus = useNameValidator(createForm, config.rulesSelector);
  const creationModal = useCreationModal();

  // Immediate validation of key name on input
  keyNameinput.addEventListener("input", () => {
    validateNameAndSetStatus(keyNameinput.value);
  });

  // submit flow
  const submitBtn = createForm.querySelector<HTMLButtonElement>(config.submitBtnSelector);
  let submitting = false;
  let createdKey: string | null = null;
  createForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (submitting) return;

    const v = validateNameAndSetStatus(keyNameinput.value);
    if (!(v.okLen && v.okChars && v.okSpaces)) {
      keyNameinput.focus();
      return
    };

    const name = v.trimmed;

    submitting = true;
    submitBtn && (submitBtn.disabled = true);

    creationModal?.showLoading();
    try {
      const response = await api.createAPIKey({ name });
      creationModal?.showSuccess(response);
      keyNameinput.value = "";
      createdKey = name;
      validateNameAndSetStatus("");
      try { config.onCreated?.(); } catch (err) { console.error("onCreated callback failed:", err); }
    } catch (err) {
      creationModal?.showError();
      console.error("Failed to create API key:", err);
    } finally {
      submitting = false;
      submitBtn && (submitBtn.disabled = false);
    }
  });
}

type RuleKey = "length" | "chars" | "spaces";

function useNameValidator(root: ParentNode, rulesSelector = DEFAULTOPTIONS.rulesSelector) {
  const rulesMap = buildRulesMap(root, rulesSelector);

  return function (raw: string) {
    const v = validateApiKeyName(raw);

    const isEmpty = v.raw.length === 0;
    const state = (ok: boolean) => (isEmpty ? "idle" : ok ? "ok" : "err") as "idle" | "ok" | "err";

    if (rulesMap.length) setValidationStatus(rulesMap.length, state(v.okLen));
    if (rulesMap.chars) setValidationStatus(rulesMap.chars, state(v.okChars));
    if (rulesMap.spaces) setValidationStatus(rulesMap.spaces, state(v.okSpaces));

    return v;
  };
}

function buildRulesMap(root: ParentNode, rulesSelector = DEFAULTOPTIONS.rulesSelector) {
  const map: Partial<Record<RuleKey, HTMLElement>> = {};
  const els = Array.from(root.querySelectorAll<HTMLElement>(rulesSelector));
  for (const el of els) {
    const key = el.dataset.validateRule as RuleKey | undefined;
    if (key) map[key] = el;
  }
  return map;
}

function setValidationStatus(el: HTMLElement, state: "idle" | "ok" | "err") {
  el.classList.remove("text-muted", "text-success", "text-danger");
  el.classList.add(state === "idle" ? "text-muted" : state === "ok" ? "text-success" : "text-danger");
}

const DEFAULTCREATIONMODALOPTIONS = {
  creationModalSelector: "#createKeyModal",
  loadingSectionSelector: "#createKeyModalLoading",
  successSectionSelector: "#createKeyModalSuccess",
  errorSectionSelector: "#createKeyModalError",
  createdApiKeyRawSelector: "#createdApiKeyRaw",
};
function useCreationModal(options?: Partial<typeof DEFAULTCREATIONMODALOPTIONS>) {
  const config = { ...DEFAULTCREATIONMODALOPTIONS, ...options };
  const modalContainer = document.querySelector(config.creationModalSelector) as HTMLElement | null;
  if (!modalContainer) return null;

  let createdKey: CreateAPIKeyData | null = null;
  const modal = useModal(modalContainer, {
    onHidden: () => {
      clearAPIKeyInSuccessSection();
      createdKey = null;
    }
  });

  const loadingContainer = modalContainer.querySelector(config.loadingSectionSelector) as HTMLElement | null;
  const successContainer = modalContainer.querySelector(config.successSectionSelector) as HTMLElement | null;
  const errorContainer = modalContainer.querySelector(config.errorSectionSelector) as HTMLElement | null;

  function showSection(which: "loading" | "success" | "error") {
    if (loadingContainer) seSectionVisibility(loadingContainer, which === "loading");
    if (successContainer) seSectionVisibility(successContainer, which === "success");
    if (errorContainer) seSectionVisibility(errorContainer, which === "error");
  }

  function setAPIKeyInSuccessSection(key: CreateAPIKeyData) {
    createdKey = key;
    if (!successContainer) return;
    renderKeyData(successContainer, createdKey.api_key as APIKey);
    const keyField = successContainer.querySelector(config.createdApiKeyRawSelector) as HTMLInputElement | null;
    if (keyField) {
      keyField.value = key.raw_api_key ?? "";
    }
  }
  function clearAPIKeyInSuccessSection() {
    if (!successContainer) return;
    clearKeyData(successContainer);
    const keyField = successContainer.querySelector(config.createdApiKeyRawSelector) as HTMLInputElement | null;
    if (keyField) {
      keyField.value = "";
    }
  }

  modalContainer.addEventListener("click", (event) => {
    const target = event.target as HTMLElement;
    const triggerCopy = target && target.closest("[data-copy-api-key]") as HTMLElement | null;
    if (triggerCopy) {
      event.preventDefault();
      copyTextWithToast(createdKey?.raw_api_key ?? "asdadas", { title: "API Key Copied", successMessage: "Your API key has been copied to clipboard." });
    }
    const triggerToggle = target && target.closest("[data-toggle-api-key]") as HTMLElement | null;
    if (triggerToggle) {
      event.preventDefault();
      const keyField = successContainer?.querySelector(config.createdApiKeyRawSelector) as HTMLInputElement | null;
      if (keyField) {
        if (keyField.type === "password") {
          keyField.type = "text";
          (triggerToggle.querySelector("i") as HTMLElement | null)?.classList.replace("fa-eye", "fa-eye-slash");
        } else {
          keyField.type = "password";
          (triggerToggle.querySelector("i") as HTMLElement | null)?.classList.replace("fa-eye-slash", "fa-eye");
        }
      }
    }
  });

  return {
    showLoading() {
      modal.setOptions({ backdrop: "static", keyboard: false });
      modal.hideHeader();
      showSection("loading");
      modal.open();
    },
    showSuccess(key: CreateAPIKeyData) {
      modal.setOptions({ backdrop: "static", keyboard: false });
      modal.setTitle("Operation successful");
      setAPIKeyInSuccessSection(key);
      showSection("success");
      modal.open();
    },
    showError() {
      modal.setOptions({ backdrop: true, keyboard: true });
      modal.setTitle("Something went wrong.");
      showSection("error");
      modal.open();
    },
  }
}

function seSectionVisibility(section: HTMLElement, visible: boolean) {
  section.style.display = visible ? "block" : "none";
  if (visible) {
    section.classList.remove("d-none");
  } else {
    section.classList.add("d-none");
  }
}