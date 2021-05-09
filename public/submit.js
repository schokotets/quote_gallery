let form = document.getElementById("form-submit");

let quotefield = document.getElementById("quotefield");
let contextfield = document.getElementById("contextfield");
let teacherselect = document.getElementById("teacherselect");
let customteacher = document.getElementsByClassName("customteacher")[0];
let customteacherfield = document.getElementById("customteacherfield");
let customteachercheckbox = document.getElementById("certainthatcustom");

let suggestionlist = document.getElementById("suggestionlist");
let confirmdifferent = document.getElementById("confirmdifferent");
let confirmdifferentcheckbox = document.getElementById(
  "confirmdifferentcheckbox"
);

let submitbtn = document.getElementById("submitbtn");
let clearformbtn = document.getElementById("clearform");
clearformbtn.addEventListener("click", clearForm());

function clearForm() {
  form.reset();
  checkTeacherSelect();
  suggestionlist.innerText = "(werden beim Schreiben geladen)";
  hideConfirmDifferent();
}

form.addEventListener("submit", processForm);

function processForm(e) {
  e.preventDefault();

  let req = {};
  req["Text"] = quotefield.value;

  let context = contextfield.value;
  if (context) {
    req["Context"] = context;
  }
  let teachervalue = teacherselect.value;
  if (teachervalue && teachervalue != " ") {
    let teacherid = parseInt(teachervalue);
    if (teacherid) {
      req["Teacher"] = teacherid;
    }
  }
  if (! ("Teacher" in req)) {
    let teachername = customteacherfield.value;
    if (teachername) {
      req["Teacher"] = teachername;
    }
  }

  axios.post("/api/quotes/submit", req)
    .then(function (res) {
      if (res.status == 200) {
        clearForm();
        alert("Erfolgreich eingesendet!");
      } else {
        return Promise.reject({ response: res });
      }
    })
    .catch(axiosErrorHandler.bind(this, "Zitat-Einsenden"));

  return true;
}

teacherselect.addEventListener("change", checkTeacherSelect);
checkTeacherSelect();

function checkTeacherSelect(e) {
  let target = e && e.target ? e.target : teacherselect;
  if (target.selectedIndex == 1) {
    //custom field
    customteacher.style.display = "unset";
    customteacherfield.setAttribute("required", true);
    customteachercheckbox.setAttribute("required", true);
  } else {
    customteacher.style.display = "none";
    customteacherfield.removeAttribute("required");
    customteachercheckbox.removeAttribute("required");
  }
}

quotefield.addEventListener("input", quoteTextInput);

let lasttimeout;
let lastquotetext;

function quoteTextInput() {
  if (lasttimeout) {
    clearTimeout(lasttimeout);
  }

  lastquotetext = quotefield.value;
  if (!quotefield.value) {
    // no quote text entered
    suggestionlist.innerText = "(werden beim Schreiben geladen)";
    return;
  }

  suggestionlist.innerText = "lädt...";
  hideConfirmDifferent();
  disableSubmitButton();
  // fetchSimilarQuotes is run after 1s of inactivity
  lasttimeout = setTimeout(fetchSimilarQuotes, 1000);
}

function hideConfirmDifferent() {
  confirmdifferent.style.display = "none";
  confirmdifferentcheckbox.removeAttribute("required");
  updateSubmitButtonState();
}

function showConfirmDifferent() {
  confirmdifferent.style.display = "unset";
  confirmdifferentcheckbox.setAttribute("required", true);
  updateSubmitButtonState();
}

quotefield.addEventListener("change", updateSubmitButtonState.bind(this, undefined, true));
teacherselect.addEventListener("change", updateSubmitButtonState);
customteacherfield.addEventListener("change", updateSubmitButtonState);
confirmdifferentcheckbox.addEventListener("change", updateSubmitButtonState);
customteachercheckbox.addEventListener("change", updateSubmitButtonState);

function updateSubmitButtonState(target, onlydisable) {
  let allrequiredfilled = true;
  for (let e of document.querySelectorAll("[required]")) {
    if (!e.value || (e.type == "checkbox" && !e.checked)) {
      allrequiredfilled = false;
      break;
    }
  }
  if (allrequiredfilled && !onlydisable) {
    enableSubmitButton();
  } else {
    disableSubmitButton();
  }
}

function disableSubmitButton() {
    submitbtn.setAttribute("disabled", "disabled");
}

function enableSubmitButton() {
    submitbtn.removeAttribute("disabled");
}

function fetchSimilarQuotes() {
  return axios
    .get("/suggestions", { params: { text: quotefield.value } })
    .then(function (res) {
      if (quotefield.value != lastquotetext) {
        //while fetching, the text has changed again, request result irrelevant
        return;
      }
      let similarQuotesHTML = res.data;
      if (similarQuotesHTML) {
        suggestionlist.innerHTML = similarQuotesHTML;
        showConfirmDifferent();
        updateSubmitButtonState();
      } else {
        suggestionlist.innerText = "Dein Zitat ist einzigartig, top :)";
        hideConfirmDifferent();
        updateSubmitButtonState();
      }
    });
}

