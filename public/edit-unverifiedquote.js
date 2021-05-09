let form = document.getElementById("form-submit");

let quotefield = document.getElementById("quotefield");
let contextfield = document.getElementById("contextfield");
let teacherselect = document.getElementById("teacherselect");

let submitbtn = document.getElementById("submitbtn");

form.addEventListener("submit", processForm);

function processForm(e) {
  e.preventDefault();

  let req = {};
  req["Text"] = quotefield.value;

  let context = contextfield.value;
  req["Context"] = context;

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

  axios.put(
    "/api/unverifiedquotes/" + window.location.pathname.split("/")[3],
    req
  ).then(function (res) {
      if (res.status == 200) {
        //hiding form because chrome re-shows last input values
        document.getElementById("form-submit").style.display = "none";
        window.location = document.referrer;
      } else {
        return Promise.reject({ response: res });
      }
    })
    .catch(axiosErrorHandler.bind(this, "Zitat-Einsenden"));

  return true;
}
