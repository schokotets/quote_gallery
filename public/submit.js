let form = document.getElementById("form-submit");

form.addEventListener("submit", processForm);

function processForm(e) {
  e.preventDefault();
  let FD = new FormData(form);

  let req = {};
  req["Text"] = FD.get("text");

  if (FD.get("context")){
    req["Context"] = FD.get("context");
  }

  if (FD.get("teacherid")){
    req["Teacher"] = parseInt(FD.get("teacherid"));
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
