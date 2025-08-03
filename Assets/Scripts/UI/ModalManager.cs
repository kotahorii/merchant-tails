using System;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using TMPro;
using MerchantTails.Core;

namespace MerchantTails.UI
{
    /// <summary>
    /// モーダルダイアログの管理システム
    /// </summary>
    public class ModalManager : MonoBehaviour
    {
        private static ModalManager instance;
        public static ModalManager Instance => instance;

        [Header("Modal Prefabs")]
        [SerializeField] private GameObject confirmModalPrefab;
        [SerializeField] private GameObject alertModalPrefab;
        [SerializeField] private GameObject inputModalPrefab;
        [SerializeField] private GameObject customModalPrefab;

        [Header("Modal Container")]
        [SerializeField] private Transform modalContainer;
        [SerializeField] private GameObject modalBackground;

        [Header("Animation Settings")]
        [SerializeField] private float showAnimationDuration = 0.3f;
        [SerializeField] private float hideAnimationDuration = 0.2f;
        [SerializeField] private AnimationCurve showAnimationCurve = AnimationCurve.EaseInOut(0, 0, 1, 1);
        [SerializeField] private AnimationCurve hideAnimationCurve = AnimationCurve.EaseInOut(0, 0, 1, 1);

        private Stack<ModalDialog> modalStack = new Stack<ModalDialog>();
        private bool isAnimating = false;

        public bool HasActiveModal => modalStack.Count > 0;
        public bool IsAnimating => isAnimating;

        private void Awake()
        {
            if (instance != null && instance != this)
            {
                Destroy(gameObject);
                return;
            }
            instance = this;

            InitializeModalSystem();
        }

        private void OnDestroy()
        {
            if (instance == this)
            {
                instance = null;
            }
        }

        private void InitializeModalSystem()
        {
            if (modalBackground != null)
            {
                modalBackground.SetActive(false);
            }
        }

        /// <summary>
        /// 確認ダイアログを表示
        /// </summary>
        public void ShowConfirmModal(
            string title,
            string message,
            Action onConfirm,
            Action onCancel = null,
            string confirmText = "はい",
            string cancelText = "いいえ"
        )
        {
            ErrorHandler.SafeExecute(() =>
            {
                if (confirmModalPrefab == null || modalContainer == null)
                    return;

                var modalGO = Instantiate(confirmModalPrefab, modalContainer);
                var modal = modalGO.GetComponent<ConfirmModal>();

                if (modal != null)
                {
                    modal.Setup(title, message, confirmText, cancelText, onConfirm, onCancel);
                    ShowModal(modal);
                }
            }, "ModalManager.ShowConfirmModal");
        }

        /// <summary>
        /// アラートダイアログを表示
        /// </summary>
        public void ShowAlertModal(string title, string message, Action onClose = null, string closeText = "OK")
        {
            ErrorHandler.SafeExecute(() =>
            {
                if (alertModalPrefab == null || modalContainer == null)
                    return;

                var modalGO = Instantiate(alertModalPrefab, modalContainer);
                var modal = modalGO.GetComponent<AlertModal>();

                if (modal != null)
                {
                    modal.Setup(title, message, closeText, onClose);
                    ShowModal(modal);
                }
            }, "ModalManager.ShowAlertModal");
        }

        /// <summary>
        /// 入力ダイアログを表示
        /// </summary>
        public void ShowInputModal(
            string title,
            string message,
            string placeholder,
            Action<string> onSubmit,
            Action onCancel = null,
            string submitText = "OK",
            string cancelText = "キャンセル"
        )
        {
            ErrorHandler.SafeExecute(() =>
            {
                if (inputModalPrefab == null || modalContainer == null)
                    return;

                var modalGO = Instantiate(inputModalPrefab, modalContainer);
                var modal = modalGO.GetComponent<InputModal>();

                if (modal != null)
                {
                    modal.Setup(title, message, placeholder, submitText, cancelText, onSubmit, onCancel);
                    ShowModal(modal);
                }
            }, "ModalManager.ShowInputModal");
        }

        /// <summary>
        /// カスタムモーダルを表示
        /// </summary>
        public void ShowCustomModal(GameObject modalPrefab)
        {
            ErrorHandler.SafeExecute(() =>
            {
                if (modalPrefab == null || modalContainer == null)
                    return;

                var modalGO = Instantiate(modalPrefab, modalContainer);
                var modal = modalGO.GetComponent<ModalDialog>();

                if (modal != null)
                {
                    ShowModal(modal);
                }
            }, "ModalManager.ShowCustomModal");
        }

        private void ShowModal(ModalDialog modal)
        {
            if (modal == null || isAnimating)
                return;

            modalStack.Push(modal);
            modal.OnCloseRequested += () => CloseModal(modal);

            // 背景を表示
            if (modalBackground != null && modalStack.Count == 1)
            {
                modalBackground.SetActive(true);
            }

            // アニメーション
            StartCoroutine(ShowModalAnimation(modal));
        }

        private void CloseModal(ModalDialog modal)
        {
            if (modal == null || isAnimating)
                return;

            if (modalStack.Count > 0 && modalStack.Peek() == modal)
            {
                modalStack.Pop();
                StartCoroutine(HideModalAnimation(modal));
            }
        }

        /// <summary>
        /// 最上位のモーダルを閉じる
        /// </summary>
        public void CloseTopModal()
        {
            if (modalStack.Count > 0 && !isAnimating)
            {
                var modal = modalStack.Peek();
                CloseModal(modal);
            }
        }

        /// <summary>
        /// すべてのモーダルを閉じる
        /// </summary>
        public void CloseAllModals()
        {
            while (modalStack.Count > 0 && !isAnimating)
            {
                var modal = modalStack.Pop();
                Destroy(modal.gameObject);
            }

            if (modalBackground != null)
            {
                modalBackground.SetActive(false);
            }
        }

        private System.Collections.IEnumerator ShowModalAnimation(ModalDialog modal)
        {
            isAnimating = true;

            var canvasGroup = modal.GetComponent<CanvasGroup>();
            if (canvasGroup == null)
            {
                canvasGroup = modal.gameObject.AddComponent<CanvasGroup>();
            }

            var rectTransform = modal.GetComponent<RectTransform>();
            if (rectTransform != null)
            {
                // スケールアニメーション
                rectTransform.localScale = Vector3.one * 0.8f;
                canvasGroup.alpha = 0f;

                float elapsed = 0f;
                while (elapsed < showAnimationDuration)
                {
                    elapsed += Time.deltaTime;
                    float t = showAnimationCurve.Evaluate(elapsed / showAnimationDuration);

                    rectTransform.localScale = Vector3.Lerp(Vector3.one * 0.8f, Vector3.one, t);
                    canvasGroup.alpha = t;

                    yield return null;
                }

                rectTransform.localScale = Vector3.one;
                canvasGroup.alpha = 1f;
            }

            isAnimating = false;
        }

        private System.Collections.IEnumerator HideModalAnimation(ModalDialog modal)
        {
            isAnimating = true;

            var canvasGroup = modal.GetComponent<CanvasGroup>();
            var rectTransform = modal.GetComponent<RectTransform>();

            if (canvasGroup != null && rectTransform != null)
            {
                float elapsed = 0f;
                while (elapsed < hideAnimationDuration)
                {
                    elapsed += Time.deltaTime;
                    float t = hideAnimationCurve.Evaluate(elapsed / hideAnimationDuration);

                    rectTransform.localScale = Vector3.Lerp(Vector3.one, Vector3.one * 0.8f, t);
                    canvasGroup.alpha = 1f - t;

                    yield return null;
                }
            }

            // モーダルを破棄
            Destroy(modal.gameObject);

            // 背景を非表示（モーダルがなくなったら）
            if (modalBackground != null && modalStack.Count == 0)
            {
                modalBackground.SetActive(false);
            }

            isAnimating = false;
        }

        private void Update()
        {
            // ESCキーでモーダルを閉じる
            if (Input.GetKeyDown(KeyCode.Escape) && HasActiveModal && !isAnimating)
            {
                var topModal = modalStack.Peek();
                if (topModal.AllowEscapeClose)
                {
                    CloseTopModal();
                }
            }
        }
    }

    /// <summary>
    /// モーダルダイアログの基底クラス
    /// </summary>
    public abstract class ModalDialog : MonoBehaviour
    {
        [Header("Base Modal Elements")]
        [SerializeField] protected TextMeshProUGUI titleText;
        [SerializeField] protected TextMeshProUGUI messageText;
        [SerializeField] protected Button closeButton;

        [Header("Modal Settings")]
        [SerializeField] private bool allowEscapeClose = true;
        [SerializeField] private bool closeOnBackgroundClick = false;

        public bool AllowEscapeClose => allowEscapeClose;
        public event Action OnCloseRequested;

        protected virtual void Awake()
        {
            if (closeButton != null)
            {
                closeButton.onClick.AddListener(RequestClose);
            }
        }

        protected virtual void OnDestroy()
        {
            if (closeButton != null)
            {
                closeButton.onClick.RemoveListener(RequestClose);
            }
        }

        protected void RequestClose()
        {
            OnCloseRequested?.Invoke();
        }

        public virtual void SetTitle(string title)
        {
            if (titleText != null)
            {
                titleText.text = title;
            }
        }

        public virtual void SetMessage(string message)
        {
            if (messageText != null)
            {
                messageText.text = message;
            }
        }
    }

    /// <summary>
    /// 確認ダイアログ
    /// </summary>
    public class ConfirmModal : ModalDialog
    {
        [Header("Confirm Modal Elements")]
        [SerializeField] private Button confirmButton;
        [SerializeField] private Button cancelButton;
        [SerializeField] private TextMeshProUGUI confirmButtonText;
        [SerializeField] private TextMeshProUGUI cancelButtonText;

        private Action onConfirmAction;
        private Action onCancelAction;

        protected override void Awake()
        {
            base.Awake();

            if (confirmButton != null)
            {
                confirmButton.onClick.AddListener(OnConfirmClicked);
            }

            if (cancelButton != null)
            {
                cancelButton.onClick.AddListener(OnCancelClicked);
            }
        }

        protected override void OnDestroy()
        {
            base.OnDestroy();

            if (confirmButton != null)
            {
                confirmButton.onClick.RemoveListener(OnConfirmClicked);
            }

            if (cancelButton != null)
            {
                cancelButton.onClick.RemoveListener(OnCancelClicked);
            }
        }

        public void Setup(
            string title,
            string message,
            string confirmText,
            string cancelText,
            Action onConfirm,
            Action onCancel
        )
        {
            SetTitle(title);
            SetMessage(message);

            if (confirmButtonText != null)
                confirmButtonText.text = confirmText;

            if (cancelButtonText != null)
                cancelButtonText.text = cancelText;

            onConfirmAction = onConfirm;
            onCancelAction = onCancel;
        }

        private void OnConfirmClicked()
        {
            onConfirmAction?.Invoke();
            RequestClose();
        }

        private void OnCancelClicked()
        {
            onCancelAction?.Invoke();
            RequestClose();
        }
    }

    /// <summary>
    /// アラートダイアログ
    /// </summary>
    public class AlertModal : ModalDialog
    {
        [Header("Alert Modal Elements")]
        [SerializeField] private Button okButton;
        [SerializeField] private TextMeshProUGUI okButtonText;

        private Action onCloseAction;

        protected override void Awake()
        {
            base.Awake();

            if (okButton != null)
            {
                okButton.onClick.AddListener(OnOKClicked);
            }
        }

        protected override void OnDestroy()
        {
            base.OnDestroy();

            if (okButton != null)
            {
                okButton.onClick.RemoveListener(OnOKClicked);
            }
        }

        public void Setup(string title, string message, string okText, Action onClose)
        {
            SetTitle(title);
            SetMessage(message);

            if (okButtonText != null)
                okButtonText.text = okText;

            onCloseAction = onClose;
        }

        private void OnOKClicked()
        {
            onCloseAction?.Invoke();
            RequestClose();
        }
    }

    /// <summary>
    /// 入力ダイアログ
    /// </summary>
    public class InputModal : ModalDialog
    {
        [Header("Input Modal Elements")]
        [SerializeField] private TMP_InputField inputField;
        [SerializeField] private Button submitButton;
        [SerializeField] private Button cancelButton;
        [SerializeField] private TextMeshProUGUI submitButtonText;
        [SerializeField] private TextMeshProUGUI cancelButtonText;

        private Action<string> onSubmitAction;
        private Action onCancelAction;

        protected override void Awake()
        {
            base.Awake();

            if (submitButton != null)
            {
                submitButton.onClick.AddListener(OnSubmitClicked);
            }

            if (cancelButton != null)
            {
                cancelButton.onClick.AddListener(OnCancelClicked);
            }

            if (inputField != null)
            {
                inputField.onSubmit.AddListener(OnInputSubmit);
            }
        }

        protected override void OnDestroy()
        {
            base.OnDestroy();

            if (submitButton != null)
            {
                submitButton.onClick.RemoveListener(OnSubmitClicked);
            }

            if (cancelButton != null)
            {
                cancelButton.onClick.RemoveListener(OnCancelClicked);
            }

            if (inputField != null)
            {
                inputField.onSubmit.RemoveListener(OnInputSubmit);
            }
        }

        public void Setup(
            string title,
            string message,
            string placeholder,
            string submitText,
            string cancelText,
            Action<string> onSubmit,
            Action onCancel
        )
        {
            SetTitle(title);
            SetMessage(message);

            if (inputField != null)
            {
                inputField.placeholder.GetComponent<TextMeshProUGUI>().text = placeholder;
                inputField.text = "";
                inputField.Select();
            }

            if (submitButtonText != null)
                submitButtonText.text = submitText;

            if (cancelButtonText != null)
                cancelButtonText.text = cancelText;

            onSubmitAction = onSubmit;
            onCancelAction = onCancel;
        }

        private void OnSubmitClicked()
        {
            if (inputField != null)
            {
                onSubmitAction?.Invoke(inputField.text);
                RequestClose();
            }
        }

        private void OnCancelClicked()
        {
            onCancelAction?.Invoke();
            RequestClose();
        }

        private void OnInputSubmit(string text)
        {
            OnSubmitClicked();
        }
    }
}