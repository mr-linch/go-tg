const WebApp = Telegram.WebApp;

WebApp.ready();

document
  .querySelector("button#backbutton")
  .addEventListener("click", (event) => {
    if (WebApp.BackButton.isVisible) {
      WebApp.BackButton.hide();
      event.currentTarget.firstChild.textContent = "BackButton.show()";
    } else {
      WebApp.BackButton.show();
      event.currentTarget.firstChild.textContent = "BackButton.hide()";
    }
  });

document.querySelector("button#expand").addEventListener("click", (event) => {
  WebApp.expand();
});

document
  .querySelector("button#mainbutton")
  .addEventListener("click", (event) => {
    if (WebApp.MainButton.text != "MAIN BUTTON") {
      WebApp.MainButton.setText("MAIN BUTTON");
    }

    if (WebApp.MainButton.isVisible) {
      WebApp.MainButton.hide();
      event.currentTarget.firstChild.textContent = "MainButton.show()";
    } else {
      WebApp.MainButton.show();
      event.currentTarget.firstChild.textContent = "MainButton.hide()";
    }
  });

document
  .querySelector("button#mainbutton-state")
  .addEventListener("click", (event) => {
    if (WebApp.MainButton.isActive) {
      WebApp.MainButton.disable();
      event.currentTarget.firstChild.textContent = "MainButton.enable()";
    } else {
      WebApp.MainButton.enable();
      event.currentTarget.firstChild.textContent = "MainButton.disable()";
    }
  });

document
  .querySelector("button#mainbutton-progress")
  .addEventListener("click", (event) => {
    if (WebApp.MainButton.isProgressVisible) {
      WebApp.MainButton.hideProgress();
      event.currentTarget.firstChild.textContent = "MainButton.showProgress()";
    } else {
      WebApp.MainButton.showProgress();
      event.currentTarget.firstChild.textContent = "MainButton.hideProgress()";
    }
  });

const hapticFeedbackImpactTypes = ["light", "medium", "heavy", "rigid", "soft"];
let hapticFeedbackImpactTypeIndex = 0;

document
  .querySelector("button#hapticfeedback-impact")
  .addEventListener("click", (event) => {
    WebApp.HapticFeedback.impactOccurred(
      hapticFeedbackImpactTypes[hapticFeedbackImpactTypeIndex]
    );

    hapticFeedbackImpactTypeIndex =
      (hapticFeedbackImpactTypeIndex + 1) % hapticFeedbackImpactTypes.length;
    event.currentTarget.firstElementChild.firstElementChild.textContent =
      hapticFeedbackImpactTypes[hapticFeedbackImpactTypeIndex];
  });

const hapticFeedNotificationTypes = ["error", "success", "warning"];
let hapticFeedNotificationTypeIndex = 0;

document
  .querySelector("button#hapticfeedback-notification")
  .addEventListener("click", (event) => {
    WebApp.HapticFeedback.notificationOccurred(
      hapticFeedNotificationTypes[hapticFeedNotificationTypeIndex]
    );

    hapticFeedNotificationTypeIndex =
      (hapticFeedNotificationTypeIndex + 1) %
      hapticFeedNotificationTypes.length;
    event.currentTarget.firstElementChild.firstElementChild.textContent =
      hapticFeedNotificationTypes[hapticFeedNotificationTypeIndex];
  });

document
  .querySelector("button#hapticfeedback-selection")
  .addEventListener("click", () => {
    WebApp.HapticFeedback.selectionChanged();
  });
