(function( $ )
{
  var listingList = document.getElementById("listingList");
  var offerFeed = document.getElementById("offerFeed");
  var offerList = document.getElementById("offerList");
  window.switchScreens = function()
  {
    if (listingList.className.indexOf("switch") > -1)
    {
      listingList.className = listingList.className.replace(/switch/gi, "");
      offerFeed.className = offerFeed.className.replace(/switch/gi, "");
    }
    else
    {
      listingList.className += " switch";
      offerFeed.className += " switch";
    }
  }

  window.onload = (function()
  {
    var loadingIcon = new Image();
    loadingIcon.src = "/img/progress.gif";

    var pager = new Pager({
      addPagerFunctionality: true,
      button: document.getElementById("offer-feed-more-btn"),
      currentPage: 0,
      pageSize: 50,
      listName: "offers",
      loadingIcon: loadingIcon,
      loadPageFunction: function(pageNum, pageSize, success, error)
      {
        $.ajax({
          url: "/webapi/offers/seller/" + pageNum,
          cache: false,
          method: "get",
          dataType: "json",
          success: success,
          error: error
        });
      },
      loadPageSuccess: function(data)
      {
        if(!data.successful)
        {
          new Dialog({
            title: "Error",
            content: "An error occurred while attempting to load a list of "+
              "offers that have been made on your listings. Please reload "+
              "the page or try again later. Sorry for the inconvenience.",
            buttons: [{text: "OK", onclick: function(){}}]
          });
          return;
        }
        var offers = data.offers;
        for(var i = 0; i < offers.length; i++)
        {
          var offer = offers[i];
          var ndOffer = document.createElement("div");
          ndOffer.id = "offer-" + offer.id;
          ndOffer.className = "feedItem";

          var ndText = document.createElement("span");
          ndText.className = "small";

          var ndBuyerLink = document.createElement("a");
          ndBuyerLink.href = "/user/profile/" + offer.buyer.username;
          ndBuyerLink.appendChild(document.createTextNode(
            offer.buyer.display_name));
          ndText.appendChild(ndBuyerLink);

          var text = " offered you $"+
            offer.price + " for ";
          ndText.appendChild(document.createTextNode(text));

          var ndListingLink = document.createElement("a");
          ndListingLink.href = "/listing/view/" + offer.listing.id;
          ndListingLink.appendChild(document.createTextNode(
            offer.listing.name));
          ndText.appendChild(ndListingLink);

          ndText.appendChild(document.createTextNode("."));

          if(offer.is_countered)
          {
            var counterText = " Your counter offer is $" +
              offer.counter + ".";
            ndText.appendChild(document.createTextNode(counterText));
          }

          if(offer.status == "accepted")
          {
            var acceptedText = " (You have accepted this offer).";
            ndText.appendChild(document.createTextNode(acceptedText));
          }

          ndOffer.appendChild(ndText);
          ndOffer.onclick = function(offer)
          {
            var sections = [];
            if(offer.buyer_comment.length > 0)
            {
              var section = {
                title: "Comments from Buyer",
                content: offer.buyer_comment
              };
              sections.push(section);
            }

            var buttons = [];

            var counterText = (offer.is_countered)?
              "Edit Counter Offer" : "Counter Offer";
            buttons.push({
              text: counterText,
              onclick: function()
              {
                window.location.href = "/offer/seller/" + offer.id;
              }
            });

            if (offer.status == "offered")
            {
              buttons.unshift({
                text: "Accept",
                onclick: function()
                {
                  offer.status = "accepted";
                  new Dialog({
                    title: "Accept Offer",
                    content: "Are you sure you would like to accept this "+
                      "offer? This will allow you to chat with the person "+
                      "who made this offer. You can reject the offer later "+
                      "if you change your mind.",
                    buttons: [
                      {text: "Yes", onclick: function()
                      {
                        var errorDialog = function()
                        {
                          new Dialog({
                            title: "Failed To Accept Offer",
                            content: "An error occurred while we were trying "+
                              "to mark the offer accepted. Please refresh the "+
                              "page or try again later.",
                            buttons: [{text: "OK", onclick: function(){}}]
                          });
                        };
                        var successDialog = function()
                        {
                          new Dialog({
                            title: "Accepted Offer",
                            content: "You have successfully accepted the "+
                              "offer. Would you like to chat with the buyer "+
                              "now?",
                            buttons: [
                              {text: "Yes", onclick: function()
                              {
                                window.location.href = "/message/client/#"+
                                  "conversation" + offer.id
                              }},
                              {text: "No", onclick: function(){}, isAlt: true}
                            ]
                          });
                        };
                        $.ajax({
                          url: "/webapi/offer/accept/" + offer.id,
                          cache: false,
                          dataType: "json",
                          data: {csrfToken: window.csrfToken},
                          success: function(data)
                          {
                            if(data.successful)
                            {
                              successDialog();
                            }
                            else
                            {
                              errorDialog();
                            }
                          },
                          error: errorDialog
                        });
                      }},
                      {text: "No", onclick: function(){}, isAlt: true}
                    ]
                  });
                }
              });
            }
            else
            {
              buttons.push({
                text: "View Conversation",
                onclick: function()
                {
                  window.location.href = "/message/client/#conversation" +
                    offer.id;
                }
              });
            }

            buttons.push({
              text: "Reject",
              onclick: function()
              {
                var rejectFn = function()
                {
                  var successFn = function()
                  {
                    var elem = document.getElementById("offer-" + offer.id);
                    elem.parentNode.removeChild(elem);
                  };
                  var errorFn = function()
                  {
                    new Dialog({
                      title: "Failed To Reject Offer",
                      content: "An unexpected error occurred and we were not "+
                        "able to mark that offer as rejected. Please try "+
                        "refreshing the page, or try again later.",
                      buttons: [{text: "OK", onclick: function(){}}]
                    });
                  }
                  $.ajax({
                    url: "/webapi/offer/delete/" + offer.id,
                    cache: false,
                    dataType: "json",
                    data: {csrfToken: window.csrfToken},
                    success: function(data)
                    {
                      if(data.successful)
                      {
                        successFn();
                      }
                      else
                      {
                        errorFn();
                      }
                    },
                    error: errorFn
                  });
                }
                new Dialog({
                  title: "Are You Sure?",
                  content: "Are you sure you would like to reject this "+
                    "offer? This action cannot be reversed.",
                  buttons: [
                    {text: "Yes", onclick: rejectFn},
                    {text: "No", onclick: function(){}, isAlt: true}
                  ]
                });
              }
            });

            new OptionPane({
              title: "Offer of $" + offer.price,
              content: "You have received an offer of $"+ offer.price+
                " from " + offer.buyer.display_name + ". How would you like "+
                "to respond to this offer?",
              sections: sections,
              buttons: buttons
            });
          }.bind(ndOffer, offer);

          offerList.appendChild(ndOffer);
        }
      },
      loadPageError: function(code)
      {
        new Dialog({
          title: "Failed To Load",
          content: "Failed to load a page of offers you have received. "+
            "Please reload the page and try again later.",
          buttons: [{ text: "OK", onclick: function(){}}]
        });
      }
    }).nextPage();
  });

  window.doListingDelete = function(id)
  {
    window.deleteListing(id, function()
    {
      var elem = document.getElementById("listing-stub-" + id);
      if(elem)
      {
        elem.parentNode.removeChild(elem);
      }
    });
  };

})( jQuery );
