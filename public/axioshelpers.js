function axiosErrorHandler(action, err) {
  if("response" in err) { // if the error is axios-generated
    alert("Fehler beim " + action + "!\n"+axiosErrorString(err.response));
  } else {
    alert("Fehler beim " + action + "!\n"+err.message);
  }
  console.error(err);
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

