import type { ApiKeysClient, APIKey } from "./api";
import { useModal } from "./modal";
import { debounce, formatDate, useValidationState, useToast, renderKeyData, clearKeyData } from "./utils";

const DEFAULTOPTIONS = {
  listContainerSelector: "#apiKeysListContainer",
  tableSelector: "#apiKeysList",
  searchSelector: "#apiKeysListSearchfield",
  deleteModalSelector: "#deleteKeyModal",
}

export function initApiKeysTable(api: ApiKeysClient, options?: Partial<typeof DEFAULTOPTIONS>) {
  const config = { ...DEFAULTOPTIONS, ...options };
  const listContainer = document.querySelector(config.listContainerSelector) as HTMLElement | null;
  const apiKeysTable = $(config.tableSelector).DataTable({
    serverSide: false,
    processing: true,
    ajax: (_req: any, callback: any, settings: any) => {
      // This hack is needed to support aborting previous requests when using custom fetch function
      let canceled = false;
      settings.jqXHR = { abort() { canceled = true; } };

      // Avoid using async/await here as for some reason it breaks DataTables' ajax handling. It will trigger unhandled rejections on aborts.
      api.getAPIKeys()
        .then((response) => {
          if (canceled) return;
          callback({ data: response.api_keys ?? [] });
          updateKeysCount(response.api_keys ?? []);
        })
        .catch((err) => {
          if (canceled) return;
          console.error("error fetching api keys", err);
          callback({ data: [] });
          updateKeysCount([]);
        });
    },
    responsive: false,
    autoWidth: false,
    paging: false,
    searching: true,
    info: false,
    lengthChange: false,
    pageLength: 0,
    columns: [
      { data: "name" },
      { data: "short_key" },
      { data: "created_at", render: (v: string) => formatDate(v) },
      { data: "last_used_at", render: (v: string) => formatDate(v) },
      { data: "disabled_at", render: (v: string) => formatDate(v) },
      {
        orderable: false, searchable: false, data: null, defaultContent: "",
        render: (_data: any, _type: any, row: APIKey) => {
          const isActiveKey = !row.disabled_at;
          return `
            <div class="d-flex ml-auto" style="gap: 0.5rem; width: 130px;">
              ${makeKeyActionBtn({
            action: isActiveKey ? 'disable' : 'enable',
            keyName: row.name,
            label: isActiveKey ? 'Disable' : 'Enable',
            class: 'text-white flex-fill'
          })}
              ${makeKeyActionBtn({
            action: 'delete',
            keyName: row.name,
            title: isActiveKey ? 'Cannot delete an active key' : 'Delete key',
            disabled: isActiveKey,
          })}
            </div>
          `;
        }
      },
    ],
    dom: "<'row'<'col-12'tr>>",
    language: {
      emptyTable: `
        <div class="text-center py-5">
          <i class="fa fa-key fa-3x text-muted" style="opacity: 0.5;"></i>
          <p class="mt-3 mb-1 h5">No API Keys</p>
          <p class="small">Create an API key and it will be listed here</p>
        </div>
      `,
    },
  });

  const input = document.querySelector(config.searchSelector) as HTMLInputElement | null;
  if (input) {
    const onType = debounce(() => {
      apiKeysTable.search(input.value).draw();
    }, 100);
    input.addEventListener("input", onType);
  }

  function updateKeysCount(api_keys: APIKey[]) {
    const totalKeysEls = document.querySelectorAll('[data-role="total-keys-count"]') as NodeListOf<HTMLElement>;
    totalKeysEls.forEach(el => {
      el.textContent = String(api_keys.length);
    });
    const activeCountEls = document.querySelectorAll('[data-role="active-keys-count"]') as NodeListOf<HTMLElement>;
    const activeKeys = api_keys.filter(k => !k.disabled_at);
    activeCountEls.forEach(el => {
      el.textContent = String(activeKeys.length);
    });
  }

  const disableKeyConfirmModal = useKeyActionConfirmModal('#disableKeyModal', {
    onConfirm: async (key) => { await api.disableAPIKey(key.name); },
    onSuccess: () => {
      apiKeysTable.ajax.reload();
    }
  });
  const deleteKeyConfirmModal = useKeyActionConfirmModal('#deleteKeyModal', {
    onConfirm: async (key) => { await api.deleteAPIKey(key.name); },
    onSuccess: () => {
      apiKeysTable.ajax.reload();
    }
  });
  const enableKey = useEnableKey({
    onConfirm: async (key) => { await api.enableAPIKey(key.name); },
    onSuccess: () => {
      apiKeysTable.ajax.reload();
    }
  });

  listContainer?.addEventListener('click', (e) => {
    const target = e.target as HTMLElement;
    const keyActionBtn = target.closest('[data-key-action]') as HTMLButtonElement | null;
    if (keyActionBtn) {
      keyActionBtn.focus(); // Ensure the button retains focus for accessibility
      const keyData = apiKeysTable.rows().data().toArray().find(k => k.name === keyActionBtn.getAttribute('data-key'));
      if (!keyData) return;
      const actionType = keyActionBtn.getAttribute('data-key-action')
      if (!isValidKeyAction(actionType)) return
      if (actionType === 'disable') disableKeyConfirmModal?.open(keyData);
      else if (actionType === 'enable') enableKey(keyData, keyActionBtn)
      else if (actionType === 'delete') deleteKeyConfirmModal?.open(keyData)
    }
  });

  return apiKeysTable;
}

function useEnableKey(options: {
  onConfirm?: (key: APIKey) => Promise<void>;
  onSuccess?: (key: APIKey) => void;
}) {
  let isEnabling = false;
  function setLoadingState(btn: HTMLElement) {
    const iconEl = btn.querySelector('i');
    if (isEnabling) {
      iconEl?.classList.remove('fa-play');
      iconEl?.classList.add('fa-spinner', 'fa-spin');
    } else {
      iconEl?.classList.remove('fa-spinner', 'fa-spin');
      iconEl?.classList.add('fa-play');
    }
  }
  return async (keyData: APIKey, triggerBtn: HTMLElement) => {
    if(isEnabling) return;
    try {
      isEnabling = true;
      setLoadingState(triggerBtn);
      await options.onConfirm?.(keyData);
      useToast({ type: "success", message: `API Key ${keyData.name} has been enabled.` });
      options.onSuccess?.(keyData);
    } catch (err) {
      console.error("Failed to enable API key:", err);
      useToast({ type: "error", message: `Failed to enable API Key ${keyData.name}. Please try again.` });
    } finally {
      isEnabling = false;
      setLoadingState(triggerBtn);
    }
  };
}


const DEFAULTKEYACTIONMODALOPTIONS = {
  confirmKeyFormSelector: "[data-key-form]",
  keyConfirmNameInputSelector: "[data-key-confirm-name]",
  keyConfirmCheckboxSelector: "[data-key-confirm-checkbox]",
  keySubmitErrorsSelector: "[data-key-submit-errors]",
  validationMissingNameMessage: "Please enter the API Key name to confirm.",
  validationNameMismatchMessage: "API Key name does not match.",
  validationCheckboxUncheckedMessage: "You must confirm to continue.",
}

function useKeyActionConfirmModal(
  modalSelector: string,
  options: Partial<typeof DEFAULTKEYACTIONMODALOPTIONS> & {
    onConfirm?: (key: APIKey) => Promise<void>;
    onSuccess?: (key: APIKey) => void;
  } = {}
) {
  const config = { ...DEFAULTKEYACTIONMODALOPTIONS, ...options };
  const modalContainer = document.querySelector(modalSelector) as HTMLElement | null;
  if (!modalContainer) return null;

  let apiKey: APIKey | null = null;

  const keyForm = modalContainer.querySelector(config.confirmKeyFormSelector) as HTMLFormElement | null;
  const keyConfirmNameInput = modalContainer.querySelector(config.keyConfirmNameInputSelector) as HTMLInputElement | null;
  const keyConfirmCheckbox = modalContainer.querySelector(config.keyConfirmCheckboxSelector) as HTMLInputElement | null;
  const keySubmitErrors = modalContainer.querySelector(config.keySubmitErrorsSelector) as HTMLElement | null;

  const keyActionConfirmModal = useModal(modalContainer, {
    backdrop: true,
    keyboard: true,
    onHidden() {
      // Need to reset data after the modal is fully hidden to avoid flickering
      apiKey = null;
      clearKeyData(modalContainer);
      keyForm?.reset();
      clearValidationState();
    },
  });

  const getErrorMessage = (key: 'validationMissingName' | 'validationNameMismatch' | 'validationCheckboxUnchecked' | 'successMessage' | 'errorMessage') =>
    modalContainer.dataset[key] || `Please check the ${key} field.`;

  let isSubmitting = false;
  keyForm?.addEventListener("submit", async (e) => {
    e.preventDefault();
    if (!apiKey || isSubmitting || !options.onConfirm) return;

    if (!validateForm()) return;

    try {
      isSubmitting = true;
      keyActionConfirmModal.showLoading();
      await options.onConfirm(apiKey);

      // Show success and close modal
      const successMessage = getErrorMessage('successMessage').replace('{keyName}', apiKey.name);
      useToast({ type: "success", message: successMessage });
      keyActionConfirmModal.close();
      options.onSuccess?.(apiKey);

    } catch (err) {
      console.error("API call failed:", err);
      const errorMessage = getErrorMessage('errorMessage').replace('{keyName}', apiKey.name);
      useToast({ type: "error", message: errorMessage });
    } finally {
      isSubmitting = false;
      keyActionConfirmModal.hideLoading();
    }
  });

  const setValidationState = useValidationState(keySubmitErrors);
  const clearValidationState = () => {
    setValidationState("", keyConfirmNameInput);
    setValidationState("", keyConfirmCheckbox);
  }
  function validateForm() {
    let isValid = true;

    clearValidationState();

    // Validate confirmation name
    const nameValue = keyConfirmNameInput?.value?.trim();
    if (!nameValue) {
      isValid = false;
      setValidationState(getErrorMessage('validationMissingName'), keyConfirmNameInput);
    } else if (nameValue !== apiKey?.name) {
      isValid = false;
      setValidationState(getErrorMessage('validationNameMismatch'), keyConfirmNameInput);
    }

    // Validate checkbox
    else if (!keyConfirmCheckbox?.checked) {
      isValid = false;
      setValidationState(getErrorMessage('validationCheckboxUnchecked'), keyConfirmCheckbox);
    }

    return isValid;
  }

  keyForm?.addEventListener("change", () => {
    validateForm();
  });

  function openKeyActionConfirmModal(keyData: APIKey, trigger?: HTMLElement) {
    apiKey = keyData;
    modalContainer && renderKeyData(modalContainer, apiKey!);
    keyActionConfirmModal.open(trigger);
  }

  return {
    open: openKeyActionConfirmModal,
  };

}


type KeyAction = 'disable' | 'enable' | 'delete';

function isValidKeyAction(action: string | null): action is KeyAction {
  return action === 'disable' || action === 'enable' || action === 'delete';
}

function makeKeyActionBtn(options: { action: KeyAction; keyName: string; label?: string; title?: string, class?: string; disabled?: boolean }): string {
  const classType = options.action !== 'delete' ? 'btn-secondary' : 'btn-danger';
  const iconType = options.action === 'disable' ? 'fa-ban' : options.action === 'enable' ? 'fa-play' : 'fa-trash';
  return `
    <button 
      type="button"
      class="btn btn-sm ${classType} ${options.class ?? ''} d-inline-flex align-items-center justify-content-center" 
      data-key-action="${options.action}" 
      data-key="${options.keyName}" ${options.disabled ? 'disabled' : ''} 
      title="${options.title ?? options.label ?? ''}" 
      ${options.disabled ? 'disabled' : ''}
      style="gap: 0.5rem;"
    >
      <i class="fa ${iconType}"></i>${options.label ?? ''}
    </button>
  `
}