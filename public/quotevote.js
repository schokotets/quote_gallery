function voteFor(button, quoteid, rating) {
  if (button.classList.contains("selected")) return;

  for (sibling of button.parentElement.children) {
    if (sibling == button) {
      sibling.classList.add("selected");
    } else {
      sibling.classList.remove("selected");
    }
  }

  axios.put("/api/quotes/" + quoteid + "/vote/" + rating);
}
