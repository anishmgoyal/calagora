(function( $ )
{

  var spinner = new Image();
  spinner.src = "/img/progress.gif";

  var ndMainImage = document.getElementById("main-image");
  var loadImage = function(image)
  {
    ndMainImage.parentNode.style.height = ndMainImage.offsetHeight + "px";
    ndMainImage.src = spinner.src;
    ndMainImage.style.position = "absolute";

    setTimeout(function()
    {
      var ndNewImage = new Image();
      ndNewImage.src = image;
      ndNewImage.onload = function()
      {
        ndMainImage.src = ndNewImage.src;
        ndMainImage.parentNode.style.height = "";
        ndMainImage.style.position = "";
      };
    }, 1);
  };

  var removeOfferTable = function()
  {
    var offerTable = document.getElementById("offerTable");
    var altButton = document.getElementById("offerButton");
    if(offerTable != null)
    {
      offerTable.parentNode.removeChild(offerTable);
    }
    if(altButton != null)
    {
      altButton.style.display = "block";
    }
  }

  var deleteListing = function(id)
  {
    new Dialog({
      title: "Are You Sure?",
      content: "Are you sure you would like to delete this listing? "+
        "This cannot be undone.",
      buttons: [
        {text: "Yes", onclick: doDeleteListing.bind(window, id)},
        {text: "No", onclick: function(){}, isAlt: true}
      ]
    });
  };

  var doDeleteListing = function(id)
  {
    var ndForm = document.createElement("form");
    ndForm.method = "POST";
    ndForm.action = "/listing/delete/" + id;

    var ndCsrfToken = document.createElement("input");
    ndCsrfToken.name = "csrfToken";
    ndCsrfToken.type = "hidden";
    ndCsrfToken.value = window.csrfToken;

    ndForm.appendChild(ndCsrfToken);
    document.body.appendChild(ndForm);
    ndForm.submit();
  };

  var offerMap = {};
  var getListingsAsSeller = function(id)
  {
    var pager = new Pager({
      addPagerFunctionality: true,
      button: document.getElementById("offer-feed-load-more"),
      currentPage: 0,
      itemList: "offers",
      pageSize: 50,
      loadingIcon: document.getElementById("offer-feed-loading"),
      loadPageFunction: function(page, pageSize, success, error) {
        $.ajax({
          url: "/webapi/listing/offers/" + id + "/" + page,
          cache: false,
          dataType: "json",
          success: success,
          error: error
        });
      },
      loadPageApi: {},
      loadPageParameters: [],
      loadPageSuccess: function(data)
      {
        if(!data.successful)
        {
          new Dialog({
            title: "Failed to Get Offers",
            content: "We weren't able to get a list of offers placed on "+
              "this listing. Try refreshing the page, or try again later. "+
              "If this issue doesn't go away, contact support at "+
              "support@calagora.com.",
            buttons: [{text: "OK", onclick: function(){}}]
          });
          return;
        }

        data = data.offers;
        if(data.length > 0)
        {
          var ndNone = document.getElementById("offer-feed-none");
          ndNone.parentNode.removeChild(ndNone);
        }

        var target = document.getElementById("offer-feed");
        for(var i = 0; i < data.length; i++)
        {
          var offer = data[i];
          offerMap[offer.id] = offer;
          var ndOffer = document.createElement("div");
          ndOffer.id = "offer-" + offer.id;
          ndOffer.className = "small small-full medium-half large-third "+
            "grid-wide";

          var ndTable = document.createElement("table");
          ndTable.className = "il";

          ndTable.appendChild(createTableRow("Buyer", offer.buyer.display_name));
          ndTable.appendChild(createTableRow("Amount", "$" + offer.price));
          if(offer.buyer_comment.length > 0)
          {
            ndTable.appendChild(createTableRow("Buyer Comments", offer.buyer_comment));
          }
          ndTable.appendChild(createTableRow("Accepted?", (offer.status == "accepted")? "Yes" : "No"));
          if(offer.is_countered)
          {
            ndTable.appendChild(createTableRow("Counter", "$" + offer.counter));
            if(offer.seller_comment.length > 0)
            {
              ndTable.appendChild(createTableRow("Your Comments", offer.seller_comment));
            }
          }

          ndOffer.appendChild(ndTable);

          if(offer.status != "accepted")
          {
            var ndAccept = document.createElement("button");
            ndAccept.id = "offer-" + offer.id + "-accept";
            ndAccept.appendChild(document.createTextNode("Accept Offer"));
            ndAccept.onclick = acceptOffer.bind(window, offer.id);
            ndOffer.appendChild(ndAccept);
          }

          var counterText = ((offer.is_countered)? "Edit " : "")+
            "Counter Offer";
          var ndCounter = document.createElement("button");
          ndCounter.id = "offer-" + offer.id + "-counter";
          ndCounter.appendChild(document.createTextNode(counterText));
          ndCounter.onclick = counterOffer.bind(window, offer.id);
          ndOffer.appendChild(ndCounter);

          if(offer.status == "accepted")
          {
            var ndViewConversation = document.createElement("button");
            ndViewConversation.id = "offer-" + offer.id + "-viewconversation";
            ndViewConversation.appendChild(document.createTextNode(
              "View Conversation"));
            ndViewConversation.onclick = viewConversation.bind(
              window, offer.id);
            ndOffer.appendChild(ndViewConversation);
          }

          var ndDelete = document.createElement("button");
          ndDelete.id = "offer-" + offer.id + "-delete";
          ndDelete.appendChild(document.createTextNode("Reject Offer"));
          ndDelete.onclick = deleteOffer.bind(window, offer.id);
          ndOffer.appendChild(ndDelete);

          target.appendChild(ndOffer);
        }
      },
      loadPageError: function()
      {
        new Dialog({
          title: "Failed to Get Offers",
          content: "We weren't able to get a list of offers placed on "+
            "this listing. Try refreshing the page, or try again later. "+
            "If this issue doesn't go away, contact support at "+
            "support@calagora.com.",
          buttons: [{text: "OK", onclick: function(){}}]
        });
      },
      useJqueryEvents: false
    });
    pager.nextPage();
  };

  function createTableRow(headerText, valueText)
  {
    var ndRow = document.createElement("tr");
    var ndHeader = document.createElement("th");
    ndHeader.appendChild(document.createTextNode(headerText));
    var ndValue = document.createElement("td");
    ndValue.appendChild(document.createTextNode(valueText));
    ndRow.appendChild(ndHeader);
    ndRow.appendChild(ndValue);
    return ndRow;
  }

  function acceptOffer(id)
  {
    var errorFn = function()
    {
      new Dialog({
        title: "Error Occurred",
        content: "An error occurred while attempting to mark that offer "+
          "accepted. Please refresh the page, or try again later. If this "+
          "issue persists, please contact support@calagora.com.",
        buttons: [{text: "OK", onclick: function(){}}]
      });
    };
    var successFn = function(data)
    {
      if(data.successful)
      {
        new Dialog({
          title: "Offer Accepted",
          content: "Now that you have accepted an offer, you can chat with "+
            "the buyer who offered it. Would you like to chat with the buyer "+
            "now? If not, you can always do this later on the Messages page.",
          buttons: [
            {text: "Yes", onclick: viewConversation.bind(window, id)},
            {text: "No", onclick: function(){}, isAlt: true}
          ]
        });
      }
      else
      {
        errorFn();
      }
    };
    new Dialog({
      title: "Accept Offer",
      content: "Are you sure you would like to accept this offer?",
      buttons: [
        {text: "Yes", onclick: function()
          {
            $.ajax({
              url: "/webapi/offer/accept/" + id,
              cache: false,
              dataType: "json",
              data: {csrfToken: window.csrfToken},
              success: successFn,
              error: errorFn
            });

            var ndOffer = document.getElementById("offer-" + id);
            var ndAccept = document.getElementById("offer-" + id + "-accept");
            var ndDelete = document.getElementById("offer-" + id + "-delete");

            var ndViewConversation = document.createElement("button");
            ndViewConversation.id = "offer-" + id + "-viewconversation";
            ndViewConversation.appendChild(document.createTextNode(
              "View Conversation"));
            ndViewConversation.onclick = viewConversation.bind(window, id);

            ndOffer.removeChild(ndAccept);
            ndOffer.insertBefore(ndViewConversation, ndDelete);
          }},
        {text: "No", onclick: function(){}, isAlt: true}
      ]
    });
  }

  function counterOffer(id)
  {
    window.location.href = "/offer/seller/" + id;
  }

  function viewConversation(id)
  {
    window.location.href = "/message/client/#conversation" + id;
  }

  function deleteOffer(id)
  {
    var errorFn = function()
    {
      new Dialog({
        title: "Error Occurred",
        content: "An error occurred while rejecting that offer. Please "+
          "refresh the page or try again later. If this issue persists, "+
          "please contact support@calagora.com.",
        buttons: [{text: "OK", onclick: function(){}}]
      });
    };

    var successFn = function(data)
    {
      if(data.successful)
      {
        var ndOffer = document.getElementById("offer-" + id);
        ndOffer.parentNode.removeChild(ndOffer);
      }
      else
      {
        errorFn();
      }
    };

    var doDelete = function()
    {
      $.ajax({
        url: "/webapi/offer/delete/" + id,
        cache: false,
        data: {csrfToken: csrfToken},
        success: successFn,
        error: errorFn
      });
    };

    new Dialog({
      title: "Are You Sure?",
      content: "Are you sure you would like to reject this offer? This "+
        "action is irreversible.",
      buttons: [
        {text: "Yes", onclick: doDelete},
        {text: "No", onclick: function(){}, isAlt: true}
      ]
    });
  }

  window.LoadImage = loadImage;
  window.DeleteListing = deleteListing;
  window.GetListingsAsSeller = getListingsAsSeller;
  window.RemoveOfferTable = removeOfferTable;

})( jQuery );
