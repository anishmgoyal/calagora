(function() {

  var transitionEndResult;
  function transitions() {
    if(transitionEndResult == undefined) {
      var list = [
        "transition",
      ];
      var div = document.createElement("div");
      for(var i = 0; i < list.length; ++i) {
        var prop = list[i];
        if(prop in div.style) {
          transitionEndResult = prop;
          return prop;
        }
      }
      transitionEndResult = false;
      return false;
    } else {
      return transitionEndResult;
    }
  }

  function websockets() {
    if(window.WebSocket) {
      return true;
    } else {
      return false;
    }
  }

  window.support = {
    transitions: transitions,
    websockets: websockets
  };

})();
