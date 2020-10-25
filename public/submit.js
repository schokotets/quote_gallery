let form = document.getElementById("form-submit");

form.addEventListener("submit", processForm);

function processForm(e) {
  e.preventDefault();

  let req = {};
  req["Text"] = document.getElementById("quotefield").value;

  let context = document.getElementById("contextfield").value;
  if (context){
    req["Context"] = context
  }
  let teacherid = document.getElementById("teacherselect").value
  if (teacherid){
    req["Teacher"] = parseInt(teacherid);
  } else {
    let teachername = document.getElementById("customteacherfield").value;
    if (teachername) {
      req["Teacher"] = teachername;
    }
  }

  axios.post("/api/quotes/submit", req)
    .then(function (res) {
      if(res.status == 200) {
        form.reset();
        alert("Erfolgreich eingesendet!");
      } else {
        return Promise.reject({response: res})
      }
    })
    .catch(function (err) {
      if("response" in err) { // if the error is axios-generated
        alert("Fehler beim Einsenden!\n"+axiosErrorString(err.response));
      } else {
        alert("Fehler beim Einsenden!\n"+err.message);
      }
      console.error(err);
    });

  return true;
}


function axiosErrorString(response) {
  if (!response) {
    return "Keine Antwort erhalten";
  }

  let errorstr = "";
  if (response.status) {
    errorstr = "Status: " + response.status;
  }
  if (response.data) {
    errorstr += "\nAntwort: " + response.data;
  }
  return errorstr;
}

let teacherselect = document.getElementById("teacherselect");
let customteacher = document.getElementsByClassName("customteacher")[0];
let customteacherfield = document.getElementById("customteacherfield");

teacherselect.addEventListener("change", checkTeacherSelect);

function checkTeacherSelect(e) {
  if(e.target.selectedIndex == 1) { //custom field
    customteacher.style.display = "unset"
    customteacherfield.setAttribute("required", true)
  } else {
    customteacher.style.display = "none"
    customteacherfield.removeAttribute("required")
  }
}
