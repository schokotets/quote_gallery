animationmap = {};

function voteFor(button, quoteid, rating) {
  if (button.classList.contains("selected")) return;

  let done = false;

  for (sibling of button.parentElement.children) {
    if (sibling == button) {
      sibling.classList.add("loading");
      animationmap[sibling.id] = true;
      sibling.classList.add("loadinganimation");
    } else {
      sibling.classList.remove("loading");
      animationmap[sibling.id] = false;
      sibling.classList.remove("selected");
    }
  }

  button.children[0].addEventListener("animationend", function (e) {
    e.target.parentElement.classList.remove("loadinganimation");
    if (animationmap[e.target.parentElement.id]) {
      setTimeout(function () {
        e.target.parentElement.classList.add("loadinganimation");
      }, 0);
    }
  });

  axios
    .put("/api/quotes/" + quoteid + "/vote/" + rating)
    .then(function (res) {
      if (res.status == 200) {
        button.classList.add("selected");
        if (res.data) {
          // vote amount in the background
          if ("Num" in res.data) {
            for (let i = 0; i < 5; i++) {
              button.parentElement.children[i].children[1].style.setProperty(
                "--score",
                res.data["Data"][i] / res.data["Num"]
              );
            }
          }
          // slider for popularity score
          if ("Pop" in res.data) {
            button.parentElement.children[5].style.setProperty(
              "--score",
              res.data["Pop"]
            );
            button.parentElement.children[5].style.setProperty(
              "display",
              "unset"
            );
          }
        }
        return Promise.resolve(res);
      } else {
        return Promise.reject(res);
      }
    })
    .catch(axiosErrorHandler.bind(this, "Abstimmen"))
    .then(function () {
      animationmap[button.id] = false;
      button.classList.remove("loading");
    });
}
