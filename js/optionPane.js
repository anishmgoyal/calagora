(function( $ )
{

  var ndTarget = document.getElementById("content");

  var transitionsEnabled = support.transitions();

  var optionPane = function optionPane_cons(params)
  {
    if(this.constructor !== optionPane_cons)
    {
      return new optionPane_cons(params);
    }

    var ndOverlay = document.createElement("div");
    ndOverlay.className = "lightOverlay";

    var ndPane = document.createElement("div");
    ndPane.className = "optionPane";

    var ndTitle = document.createElement("h3");
    ndTitle.appendChild(document.createTextNode(params.title));
    ndPane.appendChild(ndTitle);

    if(params.content)
    {
      var ndContent = document.createElement("div");
      ndContent.appendChild(document.createTextNode(params.content));
      ndContent.className = "paneContent small";
      ndPane.appendChild(ndContent);
    }

    if(params.sections)
    {
      var sections = params.sections;
      for(var i = 0; i < sections.length; i++)
      {
        var section = sections[i];
        var ndSection = document.createElement("div");
        ndSection.className = "optionPaneSection";

        var ndTitle = document.createElement("strong");
        ndTitle.appendChild(document.createTextNode(section.title));

        var ndSecCont = document.createElement("div");
        ndSecCont.appendChild(document.createTextNode(section.content));

        ndSection.appendChild(ndTitle);
        ndSection.appendChild(ndSecCont);
        ndContent.appendChild(ndSection);
      }
    }

    if(params.form)
    {
      var form = params.form;
      var ndForm = document.createElement("form");
      ndForm.className = "form";
      ndForm.action = form.action;
      ndForm.method = form.method;

      var fields = form.fields;
      for(var indField in fields)
      {
        if(fields.hasOwnProperty(indField))
        {
          var field = fields[indField];

          if(field.label)
          {
            var ndLabel = document.createElement("label");
            ndLabel.for = field.name;
            ndLabel.appendChild(document.createTextNode(field.label));
            ndForm.appendChild(ndLabel);
          }

          var ndInput = document.createElement("input");
          ndInput.type = "text";
          ndInput.placeholder = field.placeholder;
          ndInput.id = field.name;
          ndInput.name = field.name;
          if(field.autofocus)
          {
            ndInput.autofocus = "true";
          }
          ndForm.appendChild(ndInput);
        }
      }

      var ndSubmit = document.createElement("button");
      ndSubmit.type = "submit";
      ndSubmit.style.marginBottom = "0";
      ndSubmit.appendChild(document.createTextNode(form.submitText));
      ndForm.appendChild(ndSubmit);
      ndPane.appendChild(ndForm);
    }

    var buttons = params.buttons;
    if(buttons)
    {
      buttons.push({
        text: "Cancel",
        onclick: this.hide
      });

      for(var i = 0; i < buttons.length; i++)
      {
        var button = buttons[i];
        var ndButton = document.createElement("button");
        ndButton.appendChild(document.createTextNode(button.text));
        ndButton.onclick = this.coupleFunctionWithHide(button.onclick);
        ndPane.appendChild(ndButton);
      }
    }
    else
    {
      var ndLink = document.createElement("a");
      ndLink.appendChild(document.createTextNode("Cancel"));
      ndLink.className = "small";
      ndLink.href = "javascript:void(null)";
      ndLink.onclick = this.hide.bind(this);
      ndLink.style.textDecoration = "none";
      ndPane.appendChild(ndLink);
    }

    ndOverlay.appendChild(ndPane);
    ndTarget.appendChild(ndOverlay);

    this.ndOverlay = ndOverlay;

    if(transitionsEnabled)
    {
      ndOverlay.className += " opening";
      setTimeout(function()
      {
        ndOverlay.className = ndOverlay.className.replace(/opening/gi, "");
      }, 1);
    }

    this.bindClickHide();

    return this;
  };
  optionPane.prototype.bindClickHide = function()
  {
    var instance = this;
    $(this.ndOverlay).on("click", function(e)
    {
      if(e.target == instance.ndOverlay)
      {
        instance.hide();
      }
    });
  };
  optionPane.prototype.hide = function()
  {
    if(transitionsEnabled)
    {
      var $ndOverlay = $(this.ndOverlay);
      $ndOverlay.on("transitionend webkitTransitionEnd " +
        "oTransitionEnd otransitionend MSTransitionEnd", function()
      {
        $ndOverlay.remove();
      });
      this.ndOverlay.className += " closing";
      return;
    }
    this.ndOverlay.parentNode.removeChild(this.ndOverlay);
  };
  optionPane.prototype.coupleFunctionWithHide = function(fn)
  {
    if(fn == this.hide)
    {
      return fn.bind(this);
    }
    return function()
    {
      fn();
      this.hide();
    }.bind(this);
  };

  window.OptionPane = optionPane;

})( jQuery );
