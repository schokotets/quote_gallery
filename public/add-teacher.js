let form = document.getElementById("form-add");

form.addEventListener("submit", processForm);

function processForm(e) {
  e.preventDefault();

  let req = {};

  req["Title"] = document.getElementById("titlefield").value;
  req["Name"] = document.getElementById("namefield").value;
  req["Note"] = document.getElementById("notefield").value;

  let request;
  axios.post("/api/teachers", req).then(function (res) {
      if(res.status == 200) {
        form.reset();
        //hiding form because chrome re-shows last input values
        form.style.display = "none"
        window.location = document.referrer
      } else {
        return Promise.reject({response: res})
      }
    })
    .catch(function (err) {
      if("response" in err) { // if the error is axios-generated
        alert("Fehler!\n"+axiosErrorString(err.response));
      } else {
        alert("Fehler!\n"+err.message);
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
