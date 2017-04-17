(function( $ )
{

  var serverName = window.location.hostname;
  if (window.location.port.length > 0)
  {
    serverName += ":" + window.location.port;
  }
  if(window.location.protocol.indexOf("https") > -1)
  {
    serverName = "wss://" + serverName;
  }
  else
  {
    serverName = "ws://" + serverName;
  }

  function establishConnection()
  {
    if(!support.websockets())
    {
      new Dialog({
        title: "Unsupported Browser",
        content: "Your browser is not supported by Calagora. You will be able "+
          "to view listings, make offers, so on and so forth, but you will "+
          "not be able to receive real-time events (such as messages) without "+
          "refreshing your browser. We recommend the newest version of Google "+
          "Chrome, it is and always will be supported!",
        buttons: [{text: "Okay", onclick: function(){}}]
      });
      return;
    }

    window.ws = new WebSocket(serverName + "/ws/");

    ws.reconnectAttempted = false;

    ws.onopen = function(e)
    {
      ws.send(window.websockToken);
    };

    ws.onerror = function(e)
    {
      if(!ws.reconnectAttempted)
      {
        reconnectLoop();
        ws.reconnectAttempted = true;
      }
    };

    ws.onclose = function()
    {
      if(!ws.reconnectAttempted)
      {
        reconnectLoop();
        ws.reconnectAttempted = true;
      }
    };

    ws.onmessage = function(e)
    {
      var msg = e.data;
      if (msg.charAt(0) == '-')
      {
        if (msg.charAt(1) == 'I')
        {
          console.log("INFO: " + msg.substring(2));
        }
        else if(msg.charAt(1) == 'E')
        {
          console.log("ERROR: " + msg.substring(2));
        }
        else
        {
          console.log("UNEXPECTED " + msg.substring(1) + ": " +
            msg.substring(2));
        }
      }
      else
      {
        data = JSON.parse(msg);
        var value = JSON.parse(data.value);
        if (!("processNotification" in window) || !processNotification(value))
        {
          var handler = notificationHandlers[value.notif_type];
          if(handler)
          {
            handler(value.notification);
          }
        }
        addToTray(data, value);
      }
    };
  }

  var notificationHandlers = {
    "NOTIF_NEW_OFFER": function(offer)
    {
      Toast({
        content: "You received an offer of $" + offer.price + " from " +
          offer.buyer.display_name + " for " + offer.listing.name,
        link: "/listing/view/" + offer.listing.id
      });
    },
    "NOTIF_UPDATE_OFFER": function(offer)
    {
      Toast({
        content: offer.buyer.display_name + "'s offer for " +
          offer.listing.name + " was changed to $" + offer.price,
        link: "/listing/view/" + offer.listing.id
      });
    },
    "NOTIF_OFFER_COUNTER": function(offer)
    {
      Toast({
        content: offer.seller.display_name + " countered your offer for " +
          offer.listing.name + " with $" + offer.counter,
        link: "/listing/view/" + offer.listing.id
      });
    },
    "NOTIF_OFFER_REVOKED": function(offer)
    {
      Toast({
        content: offer.buyer.display_name + " revoked the offer of $" +
          offer.price + " for " + offer.listing.name + ".",
        link: "/listing/view/" + offer.listing.id
      });
    },
    "NOTIF_OFFER_REJECTED": function(offer)
    {
      Toast({
        content: offer.seller.display_name + " rejected your offer of $"+
          offer.price + " for " + offer.listing.name,
        link: "/listing/view/" + offer.listing.id
      });
    },
    "OFFER_ACCEPTED": function(offer)
    {
      Toast({
        content: offer.seller.display_name + " has accepted your offer of $"+
          offer.price + " for " + offer.listing.name,
        link: "/message/client/#conversation" + offer.id
      });
    },
    "NEW_MESSAGE": function(message)
    {
      if(message.sender.id != window.currentUser.id)
      {
        var messageText = message.message;
        if(messageText.length > 50)
        {
          messageText = messageText.substring(0, 47) + "...";
        }
        Toast({
          content: "New message from " + message.sender.display_name + ": " +
          messageText,
          link: "/message/client/#conversation" + message.offer.id
        });
      }
    }
  };

  window.acknowledgeNotifications = function(mostRecentID)
  {
    ws.send("-R" + mostRecentID);
  };

  window.acknowledgeSingleNotification = function(id)
  {
    ws.send("-r" + id);
  };

  function reconnectLoop()
  {
    new Toast({
      content: "Disconnected from Calagora. Attempting to reconnect..."
    });
    setTimeout(establishConnection, 5000);
  }

  $(document).ready(function()
  {
    establishConnection();
  });

})( jQuery );
