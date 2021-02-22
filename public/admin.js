function http(method, url) {
  axios({method: method, url: url})
    .then(function (res) {
      if(res.status == 200) {
        window.location.reload()
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

  return undefined;
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
