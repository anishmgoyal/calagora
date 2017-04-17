(function( $ )
{

  var dialog = function dialog_cons(params)
  {
    if(this.constructor !== dialog_cons)
    {
      return new dialog(params);
    }
    this.ndOverlay = this.createDialogFromParams(params);
    this.createListener();
    return this;
  };
  dialog.prototype.createDialogFromParams = function(params)
  {
    var ndOverlay = document.createElement("div");
    ndOverlay.className = "dialog-overlay";

    var ndTable = document.createElement("div");
    ndTable.className = "dialog";

    var ndHead = document.createElement("div");
    var ndHeadRow = document.createElement("div");
    var ndTitle = document.createElement("div");
    ndTitle.className = "dialog-title";
    ndTitle.appendChild(document.createTextNode(params.title));
    ndHeadRow.appendChild(ndTitle);
    ndHead.appendChild(ndHeadRow);

    var ndBody = document.createElement("div");
    var ndBodyRow = document.createElement("div");
    var ndPane = document.createElement("div");
    ndPane.className = "dialog-pane";
    ndPane.appendChild(document.createTextNode(params.content));
    ndBodyRow.appendChild(ndPane);
    ndBody.appendChild(ndBodyRow);

    var ndFoot = document.createElement("div");
    var ndFootRow = document.createElement("div");
    var cn = (params.buttons.length == 2)? "small-half" : "small-full";
    for(var i = 0; i < params.buttons.length; i++)
    {
      var button = params.buttons[i];
      var ndButton = document.createElement("div");
      ndButton.className = "dialog-button " + cn;
      if(button.isAlt)
      {
        ndButton.className += " dialog-alt";
      }
      ndButton.appendChild(document.createTextNode(button.text));
      ndButton.onclick = this.wrapClickHandler.bind(this, button.onclick);
      ndFootRow.appendChild(ndButton);
    }
    ndFoot.appendChild(ndFootRow);

    ndTable.appendChild(ndHead);
    ndTable.appendChild(ndBody);
    ndTable.appendChild(ndFoot);
    ndOverlay.appendChild(ndTable);
    document.body.appendChild(ndOverlay);

    return ndOverlay;
  };
  dialog.prototype.wrapClickHandler = function(callback)
  {
    callback();
    this.ndOverlay.parentNode.removeChild(this.ndOverlay);
  };
  dialog.prototype.keyup = function(e)
  {
    if(e.which == 13 || e.which == 27)
    {
      this.ndOverlay.parentNode.removeChild(this.ndOverlay);
      this.removeListener();
    }
  };
  dialog.prototype.click = function(e)
  {
    if(e.target.className.indexOf("dialog-overlay") != -1)
    {
      this.ndOverlay.parentNode.removeChild(this.ndOverlay);
      this.removeListener();
    }
  }
  dialog.prototype.createListener = function()
  {
    this.handler = this.keyup.bind(this);
    $(document.body).on("keyup", this.handler);
    $(this.ndOverlay).on("click", this.click.bind(this));
  };
  dialog.prototype.removeListener = function()
  {
    $(document.body).off("keyup", this.handler);
  };

  var ndToastParent = document.getElementById("navigation-bar");
  var toast = function toast_cons(params)
  {
    if(this.constructor !== toast_cons)
    {
      return new toast_cons(params);
    }

    var prevToasts = document.getElementsByClassName("toast");
    for(var i = 0, l = prevToasts.length; i < l; i++)
    {
      ndToastParent.removeChild(prevToasts[i]);
    }

    var ndToast = document.createElement("div");
    ndToast.className = "toast";
    ndToast.appendChild(document.createTextNode(params.content));

    if(params.link)
    {
      ndToast.onclick = function()
      {
        window.location.href = params.link;
      };
    }

    ndToastParent.appendChild(ndToast);
    setTimeout(ndToastParent.removeChild.bind(ndToastParent, ndToast), 5000);

    return this;
  };

  window.Dialog = dialog;
  window.Toast = toast;

})( jQuery );
