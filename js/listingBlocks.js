(function( $ )
{

  $(document).ready(function()
  {
    var elem = document.getElementsByClassName("loading-or-load-more")[0];
    if(elem)
    {
      elem.style.display = "block";
    }
  });

  if (window.sectionMode)
  {
    var listingTarget = document.getElementById("listing-list");
    var listingProgress = document.getElementById("listing-progress");
    var moreButton = document.getElementById("listing-more");
    var loadedPageCount = 0;
    var pager = new Pager({
      addPagerFunctionality: true,
      button: moreButton,
      currentPage: 0,
      pageSize: 50,
      loadingIcon: listingProgress,
      loadPageFunction: function(page, pageSize, success, error)
      {
        $.ajax({
          url: "/webapi/listings/",
          cache: false,
          dataType: "json",
          data: {
            pageNum: page,
            pageSize: pageSize,
            type: window.section
          },
          success: success,
          error: error
        });
      },
      loadPageApi: {},
      loadPageParameters: [],
      loadPageSuccess: function(data)
      {
        if(data.length > 0)
        {
          for(var i = 0; i < data.length; i++)
          {
            addListing(data[i], listingTarget);
          }
        }
        else if(loadedPageCount == 0)
        {
          addNoneFoundMessage(listingTarget);
        }
        loadedPageCount++;
      },
      loadPageError: function()
      {
        addErrorMessage(listingTarget, "information about listings currently "+
          "available");
      }
    });
    pager.nextPage();
  }
  else
  {
    var recentTarget = document.getElementById("recent-list");
    var recentProgress = document.getElementById("recent-progress");
    $.ajax({
      url: "/webapi/listings/",
      cache: false,
      dataType: "json",
      success: function(data)
      {
        if(data && data.length > 0)
        {
          for(var i = 0; i < data.length; i++)
          {
            addListing(data[i], recentTarget);
          }
        }
        else
        {
          addNoneFoundMessage(recentTarget);
        }
      },
      error: function()
      {
        addErrorMessage(recentTarget, "the most recently posted listings on "+
          "Calagora");
      },
      complete: function()
      {
        recentProgress.parentNode.removeChild(recentProgress);
      }
    });
  }

  if (window.currentUser && !window.sectionMode)
  {
    var myListingTarget = document.getElementById("my-listing-list");
    var myListingProgress = document.getElementById("my-listing-progress");

    $.ajax({
      url: "/webapi/listings/user/" + currentUser.id,
      cache: false,
      dataType: "json",
      data: {
        pageSize: 50,
        status: "listed"
      },
      success: function(data)
      {
        if(data && data.length > 0)
        {
          for(var i = 0; i < data.length; i++)
          {
            addListing(data[i], myListingTarget);
          }
        }
        else
        {
          addNoneFoundMessage(myListingTarget);
        }
      },
      error: function()
      {
        addErrorMessage(myListingTarget, "your most recently posted listings");
      },
      complete: function()
      {
        myListingProgress.parentNode.removeChild(myListingProgress);
      }
    });
  }

  function addListing(listing, target)
  {
    var ndListing = document.createElement("div");
    ndListing.className = "image-block";

    var ndListingInner = document.createElement("div");
    ndListingInner.className = "image";

    var ndImage = document.createElement("img");
    ndImage.src = listing.image_url + ".jpg";

    var ndName = document.createElement("div");
    ndName.className = "listing-name";
    ndName.appendChild(document.createTextNode(listing.name));

    var ndPlace = document.createElement("div");
    ndPlace.className = "listing-place";
    ndPlace.appendChild(document.createTextNode(listing.user.place));

    if(!listing.published)
    {
      var ndAsterisk = document.createElement("span");
      ndAsterisk.style.fontWeight = "normal";
      ndAsterisk.innerHTML = "*";
      ndName.appendChild(ndAsterisk);
    }

    var ndPrice = document.createElement("div");
    ndPrice.className = "listing-price";
    ndPrice.appendChild(document.createTextNode("$" + listing.price));

    ndListingInner.appendChild(ndImage);
    ndListing.appendChild(ndListingInner);
    ndListing.appendChild(ndName);
    if(!window.currentUser)
    {
      ndListing.appendChild(ndPlace);
    }
    ndListing.appendChild(ndPrice);

    ndListing.onclick = function()
    {
      window.location.href = "/listing/view/" + listing.id;
    };

    target.appendChild(ndListing);
  }

  function addNoneFoundMessage(target)
  {
    var ndNone = document.createElement("div");
    ndNone.className = "small";
    ndNone.appendChild(document.createTextNode(
      "No recent listings :/"));
    target.appendChild(ndNone);
  }

  function addErrorMessage(target, listOf)
  {
    new Dialog({
      title: "Failed To Load",
      content: "We were unable to get "+listOf+
        ". It is possible that you are not connected "+
        "to the internet. Please try again later.",
      buttons: [{text: "OK", onclick: function(){}}]
    });
    var ndNone = document.createElement("div");
    ndNone.className = "small";
    ndNone.appendChild(document.createTextNode(
      "An error occurred :/"));
    target.appendChild(ndNone);
  }

})( jQuery );
