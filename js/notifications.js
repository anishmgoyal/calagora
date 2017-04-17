(function( $ )
{

  if(!window.currentUser)
  {
    return;
  }

  var body = $(document.body);
  var popup = $(".notification-popup");
  var notificationToggleButton = $("#notificationOpenButton");

  var notifContainer = document.getElementsByClassName("notification-all")[0];

  var loadedPages = 0;

  var isOpen = false;
  var newestNotifID = 0;
  var loadNotificationPage = function(page)
  {
    if(!page)
    {
      page = 0;
    }

    $.ajax({
      url: "/webapi/notifications/" + page,
      dataType: "json",
      success: function(data)
      {
        var notifications = data.notifications;
        for(var i = 0, l = notifications.length; i < l; i++)
        {
          try
          {
            var notification = notifications[i];
            var value = JSON.parse(notification.value);
            var ndNotif = generateNotification(notification, value)
            notifContainer.appendChild(ndNotif);

            newestNotifID = Math.max(notification.id, newestNotifID);
          }
          catch(e)
          {}
        }
        if(l == 0)
        {
          var ndNone = document.createElement("div");
          ndNone.style.textAlign = "center";
          ndNone.innerHTML = "<br />You have no notifications.";
          ndNone.className = "notifications-none";
          notifContainer.appendChild(ndNone);
        }
      },
      error: function()
      {
        new Dialog({
          title: "Failed to Get Notifications",
          content: "For some reason, we weren't able to load notifications for "+
            "you. It's possible that you're not connected to the internet.",
          buttons: [{text: "OK", onclick: function(){}}]
        });
      }
    });
  };

  window.addToTray = function(notification, value)
  {
    var ndNotif = generateNotification(notification, value);
    if(ndNotif == null)
    {
      return;
    }

    notifContainer.insertBefore(ndNotif, notifContainer.children[0]);
    $(".notifications-none").remove();

    newestNotifID = Math.max(notification.id, newestNotifID);

    if(!window.isMessageClient && value.notif_type == "NEW_MESSAGE")
    {
      window.message_count++;
      window.renderBadges();
    }
    else if (window.isMessageClient)
    {
      if(!isOpen)
      {
        window.notification_count--;
      }
      ndNotif.className = ndNotif.className.replace(/new/g, "");
      window.acknowledgeSingleNotification(notification.id);
    }

    if(!isOpen)
    {
      window.notification_count++;
      window.renderBadges();
    }
    else
    {
      window.acknowledgeNotifications(newestNotifID);
    }
  };

  var notificationGenerators = {
    NOTIF_NEW_OFFER: function(value)
    {
      return {
        title: "New Offer",
        content: "You received a new offer of $" + value.price + " for your "+
          "listing " + value.listing.name + ".",
        link: "/listing/view/" + value.listing.id
      };
    },
    NOTIF_UPDATE_OFFER: function(value)
    {
      return {
        title: "Offer Changed",
        content: value.buyer.display_name + "'s offer on your listing "+
          value.listing.name + " was changed to $" + value.price + ".",
        link: "/listing/view/" + value.listing.id
      };
    },
    NOTIF_OFFER_COUNTER: function(value)
    {
      return {
        title: "Offer Countered",
        content: value.seller.display_name + " countered your offer of $"+
          value.price + " for " + value.listing.name + " with $"+
          value.counter + ".",
        link: "/listing/view/" + value.listing.id
      };
    },
    NOTIF_OFFER_REVOKED: function(value)
    {
      return {
        title: "Offer Revoked",
        content: value.buyer.display_name + " took back an offer of $"+
          value.price + " for your listing " + value.listing.name + ".",
        link: "/listing/view/" + value.listing.id
      };
    },
    NOTIF_OFFER_REJECTED: function(value)
    {
      return {
        title: "Offer Rejected",
        content: value.seller.display_name + " rejected your offer of $"+
          value.price + " for " + value.listing.name + ".",
        link: "/listing/view/" + value.listing.id
      };
    },
    OFFER_ACCEPTED: function(value)
    {
      return {
        title: "Offer Accepted",
        content: value.seller.display_name + " accepted your offer of $"+
          value.price + " for " + value.listing.name + ". You can now chat "+
          "with them by clicking here!",
        link: "/message/client/#conversation" + value.id
      };
    },
    NEW_MESSAGE: function(value)
    {
      if(value.sender.id != window.currentUser.id)
      {
        var message = value.message;
        if(message.length > 40)
        {
          message = message.substring(0, 37) + "...";
        }
        return {
          title: "New Message",
          content: value.sender.display_name + ": " + message,
          link: "/message/client/#conversation" + value.offer.id
        };
      }
      return null;
    }
  };

  var generateNotification = function(base, notif)
  {
    var notificationGenerator = notificationGenerators[notif.notif_type];
    if(notificationGenerator != null)
    {
      var props = notificationGenerator(notif.notification);
      if(props == null)
      {
        return null;
      }
      
      props.time = new Date(base.created);
      props.read = base.read;
      return generateNotificationNode(props);
    }
    return null;
  };

  var generateNotificationNode = function(props)
  {
    var ndNotification = document.createElement("div");
    ndNotification.className = "notification";
    if(!props.read)
    {
      ndNotification.className += " new";
    }

    var ndTitle = document.createElement("h4");
    ndTitle.className = "inline";
    ndTitle.appendChild(document.createTextNode(props.title));
    ndNotification.appendChild(ndTitle);

    var ndTimeString = document.createElement("span");
    ndTimeString.className = "small";
    var hours = props.time.getHours();
    var amPm = (hours < 12)? "am" : "pm";
    if(hours == 0)
    {
      hours = 12;
    }
    else if(hours > 12)
    {
      hours -= 12;
    }
    var minutes = props.time.getMinutes();
    if(minutes < 10)
    {
      minutes = "0" + minutes;
    }
    ndTimeString.appendChild(document.createTextNode(" " +
      (props.time.getMonth() + 1) + "/" + props.time.getDate() + "/" +
      props.time.getFullYear() + " " + hours + ":" +
      minutes + amPm));
    ndNotification.appendChild(ndTimeString);

    var ndContent = document.createElement("div");
    ndContent.appendChild(document.createTextNode(props.content));
    ndNotification.appendChild(ndContent);

    ndNotification.onclick = function(link)
    {
      window.location.href = link;
      forceCloseNotifications();
    }.bind(window, props.link);

    return ndNotification;
  }

  var forceCloseNotifications = function()
  {
    popup.removeClass("open");
    notificationToggleButton.removeClass("nav-link-active");
    body.off("click", checkForClickOutside);
    isOpen = false;
    $(".notification.new").removeClass("new");
  };

  var openNotifications = function()
  {
    popup.addClass("open");
    notificationToggleButton.addClass("nav-link-active");
    body.on("click", checkForClickOutside);
    isOpen = true;

    window.acknowledgeNotifications(newestNotifID);
    window.notification_count = 0;
    window.renderBadges();
  };

  var checkForClickOutside = function(e)
  {
    var target = e.target;
    for(var i = 0; i < 4 && target; i++)
    {
      if(target == popup[0] || target == notificationToggleButton[0])
      {
        return;
      }
      target = target.parentNode;
    }
    forceCloseNotifications();
  };

  $(document).ready(function()
  {
    notificationToggleButton.on("click", function()
    {
      if(!isOpen)
      {
        openNotifications();
      }
      else
      {
        forceCloseNotifications();
      }
    });

    $("#menuOpenButton").on("click", function()
    {
      forceCloseNotifications();
    });
    loadNotificationPage();
  });

})( jQuery );
