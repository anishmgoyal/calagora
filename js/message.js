(function( $ )
{

  window.isMessageClient = true;

  var sidebar = document.getElementById("messaging-sidebar");
  var placeholder = document.getElementById("placeholder-conversation-container");
  var main = document.getElementById("main-conversation-container");
  var messages = document.getElementById("message-list");
  var convoTitle = document.getElementById("conversation-title");

  var messageBox = document.getElementById("messageBox");
  var sendButton = document.getElementById("sendButton");

  var paddingBelow = null;

  var spinner = new Image();
  spinner.src = "/img/progress.gif";

  var convoMap = {};

  var activeConversation = null;

  function mapHashChangeListener()
  {
    if (window.location.hash.length == 0)
    {
        setActiveList();
    }

    var listener = function()
    {
      var location = window.location.hash.substring(1);
      if(location == "list")
      {
        setActiveList();
      }
      else if(location.indexOf("conversation") == 0)
      {
        var id = location.substring("conversation".length);
        setActiveConversation(id);
      }
      else
      {
        setActiveList();
      }
      $(window).trigger("resize");
    };

    window.addEventListener("hashchange", listener, false);
    listener();
  }

  function deactivateSidebarConversations()
  {
    var active = document.getElementsByClassName("messaging-sidebar-active");
    for(var i = 0; i < active.length; i++)
    {
      var elem = active[i];
      elem.className = elem.className.replace(/messaging\-sidebar\-active/g, "");
    }
  }

  var sellerButtons = document.getElementsByClassName("if-seller");
  var buyerButtons = document.getElementsByClassName("if-buyer");
  function showSellerButtons()
  {
    for(var i = 0; i < sellerButtons.length; i++)
    {
      sellerButtons[i].style.display = "";
    }
    for(i = 0; i < buyerButtons.length; i++)
    {
      buyerButtons[i].style.display = "none";
    }
  }

  function showBuyerButtons()
  {
    for(var i = 0; i < buyerButtons.length; i++)
    {
      buyerButtons[i].style.display = "";
    }
    for(i = 0; i < sellerButtons.length; i++)
    {
      sellerButtons[i].style.display = "none";
    }
  }

  function setActiveList()
  {
    main.style.display = "none";
    placeholder.style.display = "";
    main.className = main.className.replace(/active/g, "");
    sidebar.className += " active";
    deactivateSidebarConversations();
  }

  function setActiveConversation(id)
  {
    var convoBox = document.getElementById("conversation-" + id);
    if (!convoBox)
    {
      setActiveList();
      return;
    }

    activeConversation = convoMap[id];
    activeConversation.unread_count = 0;
    rerenderUnreadCount(id);

    if (currentUser.id == activeConversation.seller.id)
    {
      showSellerButtons();
    }
    else
    {
      showBuyerButtons();
    }

    deactivateSidebarConversations();
    convoBox.className += " messaging-sidebar-active";

    main.style.display = "";
    placeholder.style.display = "none";
    main.className += " active";
    sidebar.className = sidebar.className.replace(/active/g, "");
    if(id !== setActiveConversation.previous)
    {
      convoTitle.innerHTML = "";
      var ndTitle = document.createTextNode(convoMap[id].listing.name);
      convoTitle.appendChild(ndTitle);

      messages.innerHTML = "";
      paddingBelow = document.createElement("br");
      messages.appendChild(paddingBelow);

      var ndSpinner = document.createElement("div");
      ndSpinner.style.textAlign = "center";
      var imSpinner = new Image();
      imSpinner.id = "messages-spinner";
      imSpinner.src = spinner.src;
      ndSpinner.appendChild(imSpinner);
      messages.appendChild(ndSpinner);

      loadConversationPage(id, 1);

      setActiveConversation.previous = id;
    }
    else
    {
      scrollMessagesToBottom();
    }
  }
  setActiveConversation.previous = -1;

  function getConversationList()
  {
    $.ajax({
      url: "/webapi/conversation/list/",
      cache: false,
      dataType: "json",
      success: function(data)
      {
        if (data.has_error)
        {
          console.log(data.error);
          // Dialog box for error
        }

        var offers = data.offers;
        if(offers)
        {
          for(var i = 0; i < offers.length; i++)
          {
            addConversation(offers[i]);
          }
        }

        document.getElementById("instr_none_selected_progress")
          .style["display"] = "none";
        var instrId = "instr_click_to_chat";
        if (offers.length == 0)
        {
          document.getElementById("instr_none_to_display_mobile")
            .style["display"] = "block";

          instrId = "instr_none_to_display";
        }
        document.getElementById(instrId).style["display"] = "block";

        mapHashChangeListener();
      },
      error: function()
      {

      }
    });
  }

  function newMessageNotification(id)
  {
    var offer = convoMap[id];
    if (offer)
    {
      var existingNotifications =
        document.getElementsByClassName("messaging-notification");
      for(var i = 0; i < existingNotifications.length; i++) {
        var ndExisting = existingNotifications[i];
        ndExisting.parentNode.removeChild(ndExisting);
      }

      var ndNotification = document.createElement("div");
      ndNotification.className = "messaging-notification if-small";
      ndNotification.appendChild(document.createTextNode("New message in "));

      var ndListingName = document.createElement("span");
      ndListingName.className = "messaging-notification-title";
      ndListingName.appendChild(document.createTextNode(offer.listing.name));
      ndNotification.appendChild(ndListingName);

      var ndBody = document.getElementById("body");
      ndBody.appendChild(ndNotification);
      var ndParent = ndNotification.parentNode;

      ndNotification.onclick = function()
      {
        window.location.hash = "#conversation" + id;
        this.parentNode.removeChild(this);
      };

      setTimeout(ndParent.removeChild.bind(ndParent, ndNotification), 5000);
    }
  }

  function addConversation(offer)
  {
    var node = createConversationArticleNode(offer);
    sidebar.appendChild(node);
    convoMap[offer.id.toString()] = offer;
  }

  function removeConversation(offer)
  {
    var node = document.getElementById("conversation-" + offer.id);
    if (node && node.parentNode)
    {
      node.parentNode.removeChild(node);
      delete convoMap[offer.id.toString()];
    }
    if(activeConversation && activeConversation.id == offer.id)
    {
      setActiveList();
    }
  }

  function createConversationArticleNode(offer)
  {
    var link = document.createElement("a");
    link.href = "#conversation" + offer.id;

    var article = document.createElement("article");
    article.id = "conversation-" + offer.id;
    article.className = "messaging-sidebar-conversation";

    var ndTitle = document.createElement("h5");
    ndTitle.appendChild(document.createTextNode(offer.listing.name));

    var otherUser = offer.seller;
    if (currentUser.id == offer.seller.id)
    {
      otherUser = offer.buyer;
    }

    var ndPriceAndName = document.createElement("div");
    ndPriceAndName.className = "small";
    ndPriceAndName.appendChild(
      document.createTextNode(
        "$" + offer.price + " (" + otherUser.display_name + ")"));

    article.appendChild(ndTitle);
    article.appendChild(ndPriceAndName);

    if (offer.unread_count > 0) {
      var ndUnreadCount = document.createElement("div");
      var plural = (offer.unread_count > 1)? "S" : "";

      ndUnreadCount.className = "messaging-sidebar-unread-count";
      ndUnreadCount.appendChild(
        document.createTextNode(offer.unread_count + " UNREAD MESSAGE" + plural));
      article.appendChild(ndUnreadCount);
    }

    link.appendChild(article);
    return link;
  }

  function loadConversationPage(id, page)
  {
    id = id * 1;
    $.ajax({
      url: "/webapi/messages/" + id + "/" + page,
      cache: false,
      dataType: "json",
      success: function(data)
      {
        var ndSpinner = document.getElementById("messages-spinner");
        if(ndSpinner)
        {
          ndSpinner.parentNode.removeChild(ndSpinner);
        }

        if (data.has_error)
        {
          console.log(data.error);
          return;
        }

        var newMessages = data.messages;
        var prevFront = messages.childNodes[0];
        for(var i = newMessages.length - 1; i >= 0; i--)
        {
          addMessageBefore(newMessages[i], prevFront, id);
        }

        if(page == 1)
        {
          scrollMessagesToBottom();
        }
      },
      error: function()
      {
        console.log("Failed to load messages");
      }
    });
  }

  function addMessageBefore(message, prevFront, id)
  {
    var node = createMessageNode(message);
    if(activeConversation &&
      (window.location.hash == "#conversation" + id ||
      (window.location.hash == "#list" && activeConversation.id == id)))
    {
      messages.insertBefore(node, prevFront);
    }
  }

  function updateUnreadCount(message)
  {
    if(window.location.hash == "#conversation" + message.offer.id)
    {
      $.ajax({
        url: "/message/read/" + message.id
      });
    }
    else
    {
      if(window.location.hash != "#list")
      {
        newMessageNotification(message.offer.id);
      }
      convoMap[message.offer.id].unread_count ++;
      rerenderUnreadCount(message.offer.id);
    }
  }

  function rerenderUnreadCount(id)
  {
    var unreadCount = convoMap[id].unread_count;
    var plural = (unreadCount > 1)? "S" : "";
    var ndConvoBox = document.getElementById("conversation-" + id);
    if (ndConvoBox)
    {
      var ndUnreadCount = ndConvoBox.childNodes[ndConvoBox.childNodes.length - 1];
      if (ndUnreadCount.className.indexOf("messaging-sidebar-unread-count") == -1)
      {
        if(unreadCount > 0)
        {
          ndUnreadCount = document.createElement("div");
          ndUnreadCount.className = "messaging-sidebar-unread-count";
          ndConvoBox.appendChild(ndUnreadCount);
        }
      }
      else if(unreadCount == 0)
      {
        ndUnreadCount.parentNode.removeChild(ndUnreadCount);
      }
      if (unreadCount > 0) {
        ndUnreadCount.innerHTML = "";
        ndUnreadCount.appendChild(document.createTextNode(
            convoMap[id].unread_count + " UNREAD MESSAGE" + plural
        ));
      }
    }
  }

  function createMessageNode(message)
  {
    var ndMessage = document.createElement("div");
    ndMessage.className = "conversation-message";

    var isSender = false;
    if (message.sender.id == currentUser.id)
    {
      isSender = true;
      ndMessage.className += " sender-self";
    }

    var ndSender = document.createElement("span");
    ndSender.className = "conversation-message-sender";
    var senderText = "You";
    if(!isSender)
    {
      senderText = message.sender.display_name;
    }
    var textSender = document.createTextNode(senderText);
    ndSender.appendChild(textSender);

    var ndTimestamp = document.createElement("span");
    ndTimestamp.className = "conversation-message-timestamp small";
    var timestampDate = new Date(message.created);
    var textTimestamp = document.createTextNode(dateString(timestampDate));
    ndTimestamp.appendChild(textTimestamp);

    var ndMessageText = document.createElement("div");
    var textMessageText = document.createTextNode(message.message);
    ndMessageText.appendChild(textMessageText);

    ndMessage.appendChild(ndSender);
    ndMessage.appendChild(document.createTextNode(" "));
    ndMessage.appendChild(ndTimestamp);
    ndMessage.appendChild(document.createTextNode(" "));
    ndMessage.appendChild(ndMessageText);
    return ndMessage;
  }

  function sendMessageIfEnter(e)
  {
    e = e || window.event;
    var code = (typeof e.which == "number")? e.which : e.keyCode;
    if (code == 13)
    {
      sendMessage();
      if (e.preventDefault)
      {
        e.preventDefault();
      }
      else
      {
        return false;
      }
    }
  }

  function sendMessage()
  {
    var message = messageBox.value;
    messageBox.value = "";

    if (message.length == 0)
    {
      return;
    }

    var error_func = function()
    {
        new Dialog({
          title: "Failed to Send",
          content: "Your message, \"" + message + "\" failed to send. "+
            "It is possible that this is because you are not connected to the "+
            "internet, or because this offer has already been deleted.",
          buttons: [{text: "Got It", onclick: function() {}}]
        })
    };

    $.ajax({
      url: "/webapi/message/send/" + activeConversation.id,
      cache: false,
      data: {
        message: message,
        csrfToken: window.csrfToken
      },
      dataType: "json",
      success: function(data)
      {
        if(data.has_error)
        {
          error_func();
        }
      },
      error: function()
      {
        error_func();
      }
    });
  }

  window.processNotification = function(msg)
  {
    if (msg.notif_type == "NEW_MESSAGE")
    {
      var notif = msg.notification;

      notif.created = new Date();
      addMessageBefore(notif, paddingBelow, notif.offer.id);
      updateUnreadCount(notif);
      scrollMessagesToBottom();

      return true;
    }
    return false;
  };

  window.editOffer = function()
  {
    if (activeConversation)
    {
      window.location = "/offer/buyer/" + activeConversation.listing.id;
    }
  };

  window.deleteOffer = function(name)
  {
    var id = activeConversation.id;
    new Dialog({
      title: name + " Offer",
      content: "Are you sure you would like to " + name.toLowerCase() +
        " this offer? This action cannot be reversed.",
      buttons: [
        {text: "Yes", onclick: function()
          {
            var errorDialog = function()
            {
              new Dialog({
                title: "Failed To Delete Offer",
                content: "The offer could not be deleted due to an unexpected " +
                "error. Please try again later or contact support at "+
                "support@calagora.com",
                buttons: [{text: "OK", onclick: function() {}}]
              });
            };
            $.ajax({
              url: "/webapi/offer/delete/" + id,
              cache: false,
              data: {
                csrfToken: window.csrfToken
              },
              success: function(data)
              {
                if(data.successful)
                {
                  removeConversation({id: id});
                }
                else
                {
                  errorDialog();
                }
              },
              error: function()
              {
                errorDialog();
              }
            });
          }},
        {text: "No", onclick: function(){}, isAlt: true}
      ]
    });
  };

  window.finalizeOffer = function()
  {
    var id = activeConversation.id;
    new Dialog({
      title: "Finalize Transaction",
      content: "Are you sure you would like to finalize this transaction? This " +
        "will delete any associated messages, mark your listing sold, and remove " +
        "your listing from the Calagora.",
      buttons: [
        {text: "Yes", onclick: function()
          {
            var errorDialog = function()
            {
              new Dialog({
                title: "Failed To Finalize Transaction",
                content: "The transaction could not be finalized due to an "+
                "unexpected error. Please try again later or contact support at "+
                "support@calagora.com",
                buttons: [{text: "OK", onclick: function() {}}]
              });
            };
            $.ajax({
              url: "/webapi/offer/finalize/" + id,
              cache: false,
              data: {
                csrfToken: window.csrfToken
              },
              success: function(data)
              {
                if(data.successful)
                {
                  removeConversation({id: id});
                }
                else
                {
                  errorDialog();
                }
              },
              error: function()
              {
                errorDialog();
              }
            })
          }},
        {text: "No", onclick: function(){}, isAlt: true}
      ]
    });
  };

  function scrollMessagesToBottom()
  {
    //$(messages).scrollTop($(paddingBelow).offset().top + messages.scrollHeight);
    setTimeout(function()
    {
      $(messages).scrollTop(messages.scrollHeight);
    }, 10);
  };

  // startup scripts

  sendButton.onclick = sendMessage;
  messageBox.onkeydown = sendMessageIfEnter;
  messageBox.onfocus = scrollMessagesToBottom;

  getConversationList();

  var supportContent = $("#content");
  var supportContentBody = $(".content-body");
  var supportMessages = $(messages);
  var supportRowTop = $("#conversation-toprow");
  var supportRowButton = $("#conversation-buttonrow");
  var supportRowBottom = $("#conversation-bottomrow");
  function messagingSupport()
  {
    setTimeout(function() {
      supportContentBody.css("height", "");
      supportMessages.css("height", "");
      if(supportContentBody.outerHeight() < supportContent.innerHeight())
      {
        // We may have a height problem, fix it
        var totalHeight = supportContent.innerHeight();
        supportContentBody.css("height", (totalHeight) + "px");

        setTimeout(function() {
          var totalPadding = supportRowTop.outerHeight() +
            supportRowButton.outerHeight() +
            supportRowBottom.outerHeight();

          supportMessages.css("height", (totalHeight-totalPadding) + "px");
        }, 1);

        $(document).ajaxSuccess(messagingSupport);
      }
    }, 1);
  }

  $(window).resize(messagingSupport);
  $(document).ready(messagingSupport());

})( jQuery );
