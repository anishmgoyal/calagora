(function( $ )
{
  window.deleteOffer = function(id, verb, successCallback)
  {
    var successFn = (successCallback)? successCallback : function(){};
    var errorFn = function()
    {
      new Dialog({
        title: "Failed To " + verb + " Offer",
        content: "We weren't able to process your request. Please try "+
          "refreshing the page, or try again later.",
        buttons: [{text: "OK", onclick: function(){}}]
      });
    };

    var doDelete = function()
    {
      $.ajax({
        url: "/webapi/offer/delete/" + id,
        cache: false,
        data: {
          csrfToken: window.csrfToken
        },
        dataType: "json",
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
      })
    };

    new Dialog({
      title: verb + " Offer",
      content: "Are you sure you want to " + verb.toLowerCase() + " this "+
        "offer? This action cannot be undone.",
      buttons: [
        {text: "Yes", onclick: doDelete},
        {text: "No", onclick: function(){}, isAlt: true}
      ]
    });
  };

  window.deleteListing = function(id, successCallback)
  {
    var successFn = (successCallback)? successCallback : function(){};
    var errorFn = function()
    {
      new Dialog({
        title: "Failed to Delete Listing",
        content: "We weren't able to delete that listing. Please try "+
          "refreshing the page, or try again later.",
        buttons: [{text: "OK", onclick: function(){}}]
      });
    };

    var doDelete = function()
    {
      $.ajax({
        url: "/webapi/listing/delete/" + id,
        cache: false,
        data: {
          csrfToken: window.csrfToken
        },
        dataType: "json",
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
    };

    new Dialog({
      title: "Delete Listing",
      content: "Are you sure you would like to delete this listing? This "+
      "action cannot be reversed.",
      buttons: [
        {text: "Yes", onclick: doDelete},
        {text: "No", onclick: function(){}, isAlt: true}
      ]
    });
  };

})( jQuery );
