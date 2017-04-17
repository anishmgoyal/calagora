if(!window.jQuery)
{
  window.location.href= "/unsupported";
}

(function( $ ) {

  function setupSidebarEvents()
  {
    var menuOpenButton = $("#menuOpenButton");
    var menuCloseButton = $("#menuCloseButton");
    var sidebar = $("#sidebar");
    var body = $(document.body).add(document.documentElement);

    var prevSidebarScrollTop = 0;

    var transitionsEnabled = support.transitions();

    menuOpenButton.on("click", function()
    {
      var openClass = (transitionsEnabled)? "content-sidebar-opening" :
                          "content-sidebar-open";
      sidebar.removeClass("content-sidebar-closing")
        .addClass(openClass);

      if(!transitionsEnabled)
      {
        prevSidebarScrollTop = body.scrollTop();
        body.addClass("menuOpen");
      }
    });

    menuCloseButton.on("click", function()
    {
      sidebar.removeClass("content-sidebar-opening")
        .removeClass("content-sidebar-open")
      if(transitionsEnabled)
      {
        sidebar.addClass("content-sidebar-closing");
      }

      body.removeClass("menuOpen");
      setTimeout(function() { body.scrollTop(prevSidebarScrollTop); }, 1);
    });

    sidebar.on("transitionend webkitTransitionEnd " +
      "oTransitionEnd otransitionend MSTransitionEnd", function()
    {
      if(sidebar.hasClass("content-sidebar-opening"))
      {
        sidebar.removeClass("content-sidebar-opening")
          .addClass("content-sidebar-open");

        prevSidebarScrollTop = body.scrollTop();
        body.addClass("menuOpen");
      }
      else if(sidebar.hasClass("content-sidebar-closing"))
      {
        sidebar.removeClass("content-sidebar-closing");
      }
    });

    var dropdowns = sidebar.find(".content-sidebar-has-dropdown > a");
    dropdowns.on("click", function(e)
    {
      if(sidebar.hasClass("content-sidebar-open"))
      {
        var openClass = (transitionsEnabled)? "content-dropdown-opening" :
                            "content-dropdown-open";

        var instance = $(this.parentNode);
        instance.parent().scrollTop(0);
        instance.find(".content-sidebar-dropdown")
          .addClass(openClass);

        if(!transitionsEnabled)
        {
            sidebar.addClass("content-sidebar-has-open-dropdown");
        }
      }
    });

    dropdowns = sidebar.find(".content-sidebar-dropdown");
    dropdowns.on("transitionend webkitTransitionEnd " +
      "oTransitionEnd otransitioned MSTransitionEnd", function()
      {
        var instance = $(this);
        if(instance.hasClass("content-dropdown-opening"))
        {
          instance.removeClass("content-dropdown-opening")
            .addClass("content-dropdown-open");
          sidebar.addClass("content-sidebar-has-open-dropdown");
        }
        else if (instance.hasClass("content-dropdown-closing"))
        {
          instance.removeClass("content-dropdown-closing");
        }
      });
    dropdowns.each(function()
    {
      var instance = $(this);
      var closeButton = instance.find(".content-sidebar-close-button");
      closeButton.on("click", function(e)
      {
        if(instance.hasClass("content-dropdown-open"))
        {
          instance.removeClass("content-dropdown-open");
          if(transitionsEnabled)
          {
            instance.addClass("content-dropdown-closing");
          }
          sidebar.removeClass("content-sidebar-has-open-dropdown");
          e.preventDefault();
        }
      });
    });
    $(".content-sidebar-has-dropdown > a").on("click", function(e)
    {
      var instance = $(this);
      var dropdown = instance.parent().find(".content-sidebar-dropdown");
      if(!dropdown.hasClass("content-sidebar-dropdown-focus"))
      {
        $(".content-sidebar-dropdown-focus")
          .removeClass("content-sidebar-dropdown-focus");

        dropdown.addClass("content-sidebar-dropdown-focus");
        body.off("click", setupSidebarEvents.checkForUnfocus);
        setTimeout(body.on.bind(body, "click",
          setupSidebarEvents.checkForUnfocus), 1);
      }
    });
  }
  setupSidebarEvents.checkForUnfocus = function(e)
  {
    var targ = $(e.target);
    var body = $(document.body).add(document.documentElement);
    do {
      if(targ.hasClass("content-sidebar-dropdown-focus"))
      {
        return;
      }
    } while((targ = targ.parent()).length > 0);
    $(".content-sidebar-dropdown-focus")
      .removeClass("content-sidebar-dropdown-focus");
    body.off("click", setupSidebarEvents.checkForUnfocus);
  };

  window.dateString = function(d)
  {
    var month = d.getMonth() + 1;
    var date = d.getDate();
    var year = d.getFullYear();

    var hour = d.getHours();
    var ampm = "";
    if (hour == 0)
    {
      hour = "12";
      ampm = "am";
    }
    else if(hour < 12)
    {
      if (hour < 10)
      {
        hour = "0" + hour;
      }
      else
      {
        hour = hour.toString();
      }
      ampm = "am";
    }
    else if(hour == 12)
    {
      hour = "12";
      ampm = "pm";
    }
    else
    {
      hour = hour - 12;
      if (hour < 10)
      {
        hour = "0" + hour;
      }
      else
      {
        hour = hour.toString();
      }
      ampm = "pm";
    }
    var minute = d.getMinutes();
    if (minute < 10)
    {
      minute = "0" + minute;
    }
    else
    {
      minute = minute.toString();
    }
    return month + "/" + date + "/" + year + " " + hour + ":" + minute + ampm;
  }

  window.openSearchWindow = function()
  {
    new OptionPane({
      title: "Search Calagora",
      form: {
        action: "/search/",
        method: "get",
        submitText: "Search",
        fields: [
          {
            name: "q",
            placeholder: "Enter your search here",
            autofocus: false
          }
        ]
      }
    });

    setTimeout(function()
    {
      var ndQ = document.getElementById("q");
      if(ndQ)
      {
        ndQ.click();
        ndQ.focus();
        ndQ.ontouchstart = function() { ndQ.focus(); };
        $(ndQ).trigger('touchstart');
      }
    }, 1);
  };

  window.openProfileWindow = function()
  {
    if(window.currentUser)
    {
      new OptionPane({
        title: "Profile",
        content: "You are currently logged in as "+
          currentUser.display_name,
        buttons: [
          {text: "Edit Profile", onclick: function(){
            window.location.href = "/user/profile";
          }},
          {text: "Logout", onclick: function(){
            window.location.href = "/user/logout";
          }}
        ]
      });
    }
    else
    {
      window.location.href = "/user/login/";
    }
  };

  window.renderBadges = function()
  {
    var messageCount = (window.message_count == 0)?
      "" : (window.message_count > 99)? "+" : window.message_count;
    var notificationCount = (window.notification_count == 0)?
      "" : (window.notification_count > 99)? "+" : window.notification_count;

    if(!window.isMessageClient)
    {
      window.message_icon.attr("data-badge", messageCount);
    }
    window.notification_icon.attr("data-badge", notificationCount);
  };

  $(document).ready(function()
  {
    setupSidebarEvents();

    if(window.currentUser)
    {
      $.ajax({
        url: "/webapi/notification/counts/",
        dataType: "json",
        success: function(data)
        {
          window.message_count = data.message_count;
          window.message_icon = $(".fi-mail");
          window.notification_count = data.notification_count;
          window.notification_icon = $(".fi-sound");
          renderBadges();
        }
      });
    }
  });

  var supportContent = $("#content");
  var supportSidebar = $(".content-sidebar");
  var supportBody = $(".content-body");
  function sidebarSupport()
  {
    supportSidebar.css("height", "");
    setTimeout(function() {
      var sidebarHeight = supportSidebar.height();
      if(sidebarHeight < 10)
      {
        var totalHeight = Math.max(supportContent.height(), supportBody.height());
        supportSidebar.css("height", (totalHeight+1) + "px");
      }
    }, 1);

    $(document).ajaxSuccess(sidebarSupport);
  }

  $(window).resize(sidebarSupport);
  $(document).ready(sidebarSupport);

})( jQuery );
