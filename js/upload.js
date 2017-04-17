(function( $ )
{

  var activeUpload = false;
  var totalSize = 0;

  var imageCount = 0;

  var progressBar = document.getElementById("progress");

  var instructionPane = document.getElementById("imageUploaderDefault");
  var progressPane = document.getElementById("imageUploaderProgress");

  var allowedTypes = {
    "image/gif": true,
    "image/jpeg": true,
    "image/jpg": true,
    "image/png": true
  };

  var processingImages = {};
  function addToProcessing(images)
  {
    for(var i = 0; i < images.length; i++)
    {
      createProcessingBlock(images[i].name);
    }
  }

  var processingBlock = document.getElementById("processingBlock");
  function createProcessingBlock(name)
  {
    if(document.getElementById("processing-im-" + name))
    {
      return;
    }
    var ndBlock = document.createElement("div");
    ndBlock.id = "processing-im-" + name;
    ndBlock.appendChild(document.createTextNode("Processing " + name));
    processingBlock.appendChild(ndBlock);
  }

  function removeProcessingBlock(name)
  {
    var elem = document.getElementById("processing-im-" + name);
    if(elem)
    {
      elem.parentNode.removeChild(elem);
    }
  }

  function uploadFiles(listing_id)
  {
    if(activeUpload)
    {
      new Dialog({
        title: "Upload In Progress",
        content: "Please wait for the current upload to finish before uploading "+
          "more images",
        buttons: [{text: "OK", onclick: function(){}}]
      });
      return false;
    }

    var uploadForm = document.getElementById("uploadForm");
    var iframe = document.createElement("iframe");
    var id = "upload_iframe_" + uploadFiles.numUploads++;

    iframe.id = id;
    iframe.name = id;
    iframe.style.display = "none";

    var files = document.getElementById("upload");

    totalSize = 0;
    if (files.files)
    {
      if (files.files.length > 8 - imageCount)
      {
        new Dialog({
          title: "Too Many Files",
          content: "Each listing can only have up to 8 images",
          buttons: [{text: "OK", onclick: function(){}}]
        });
        return false;
      }

      files = files.files;
      for(var i = 0, l = files.length; i < l; i++)
      {
        var file = files[i];
        totalSize += file.size;
        if(file.size > 5242880)
        {
          new Dialog({
            title: "Upload Error",
            content: "The image \"" + file.name + "\" is larger than 5MB, so "+
              "it can't be used as an image for your listing.",
            buttons: [{text: "OK", onclick: function(){}}]
          });
          return false;
        }
        else if(!allowedTypes.hasOwnProperty(file.type))
        {
          new Dialog({
            title: "Upload Error",
            content: "The file \"" + file.name + "\" is not a GIF, JPEG/JPG, "+
              "or PNG file, so it can't be used as an image for your listing.",
            buttons: [{text: "OK", onclick: function(){}}]
          });
          return false;
        }
      }
    }
    else
    {
      new Dialog({
        title: "Unsupported Browser",
        content: "To upload files, please use a newer browser, such "+
          "as the lastest version of Google Chrome or Mozilla Firefox.",
        buttons: [{text: "OK", onclick: function(){}}]
      });
      return false;
    }

    uploadForm.target = id;
    document.body.appendChild(iframe);

    var token = window.uploadTokenStem.replace(/\//g, "_") +
      (new Date().getTime() % 1000000);
    var echoSelf = uploadFiles.numUploads + "_" +
      ((new Date().getTime() * uploadFiles.numUploads) % 1000000);

    listing_id = encodeURIComponent(listing_id);
    token = encodeURIComponent(token);
    echoSelf = encodeURIComponent(echoSelf);

    uploadForm.action = "/upload/" + listing_id + "/" + window.csrfToken +
      "/" + token + "/" + echoSelf;
    uploadForm.submit();
    iframe.onload = function()
    {
      activeUpload = false;
      instructionPane.style.display = "block";
      progressPane.style.display = "none";

      var $iframe = $(iframe);
      try {
        var response = JSON.parse($iframe.contents().find("body").text());
        if (response.successful)
        {
          addToProcessing(response.images);
          if(response.failed_images.length > 0)
          {
            var imageList = "";
            for(var i = 0; i < repsonse.failed_images.length; i++)
            {
              if(imageList.length > 0)
              {
                imageList += ", ";
              }
              imageList += response.failed_images[i];
            }
            new Dialog({
              title: "Images Failed To Upload",
              content: "The following images failed to upload: " + imageList,
              buttons: [{text: "OK", onclick: function(){}}]
            });
          }
        }
        else
        {
          console.log(response);
          new Dialog({
            title: "Upload Failed",
            content: "Your upload failed due to an unexpected error.",
            buttons: [{text: "OK", onclick: function(){}}]
          });
        }
      } catch(e) {
        new Dialog({
          title: "Failed To Fetch Upload Result",
          content: "Your browser is unable to determine if the images "+
            "were uploaded successfully. Please refresh the page to "+
            "see if your images were uploaded successfully.",
          buttons: [{text: "OK", onclick: function(){}}]
        });
      }
    };
    activeUpload = true;
    instructionPane.style.display = "none";
    progressPane.style.display = "block";

    switchToIndeterminate();
    getUploadProgress(token, echoSelf);

    return true;
  }
  uploadFiles.numUploads = 0;

  function getUploadProgress(token, echoSelf)
  {
    if(!activeUpload)
    {
      return;
    }

    $.ajax({
      url: "/webapi/upload/progress/" + token,
      cache: false,
      method: "get",
      dataType: "json",
      success: function(data)
        {
          var t = setTimeout(
            getUploadProgress.bind(window, token, echoSelf), 10);

          if(!data.successful || data.upload_progress.echo_self != echoSelf)
          {
            switchToIndeterminate();
            return;
          }
          var read = 0;
          var files = data.upload_progress.files;
          for(var param in files)
          {
            if(files.hasOwnProperty(param))
            {
              read += files[param].bytes_read;
            }
          }
          var progress = read * 100.0 / totalSize;
          if(progress > 0)
          {
            switchToDeterminate();
            progressBar.style.width = progress.toFixed(2) + "%";
          }
          if(progress >= 100)
          {
            clearTimeout(t);
          }
        },
      error: function()
        {
          switchToIndeterminate();
          return;
        }
    });
  }

  var isDeterminate = true;
  function switchToIndeterminate()
  {
    if(isDeterminate)
    {
      progressBar.style.width = "100%";
      progressBar.className += " indeterminate";
      isDeterminate = false;
    }
  }
  function switchToDeterminate()
  {
    if(!isDeterminate)
    {
      progressBar.style.width = 0;
      progressBar.className = progressBar.className.replace(
        /indeterminate/gi, "");
      isDeterminate = true;
    }
  }

  var uploadField = document.getElementById("upload");
  uploadField.onchange = function(e)
  {
    var listing_id = document.getElementById("listing_id").value;
    if(!uploadFiles(listing_id))
    {
      activeUpload = false;
      if(e.preventDefault)
      {
        e.preventDefault();
      }
      else
      {
        return false;
      }
    }
  };

  var uploader = document.getElementById("imageUploader");
  uploader.onclick = function(e)
  {
    uploadField.click();
  };

  var thumbnailDiv = document.getElementById("listingImageThumbnails");

  var images = window.images;
  for(var i = 0; i < images.length; i++)
  {
    var image = images[i];
    if (image.url.length > 0)
    {
      addImage(image);
    }
    else
    {
      addToProcessing([image]);
    }
  }

  function addImage(image)
  {
    imageCount++;
    var ndImageBlock = document.createElement("div");
    ndImageBlock.className = "image-block";
    ndImageBlock.id = "thumbnail-" + image.id;

    var ndLink = document.createElement("a");
    ndLink.appendChild(document.createTextNode("Delete"));
    ndLink.href = "javascript:void(null)";
    ndLink.onclick = function(id)
    {
      deleteImage(id);
    }.bind(window, image.id);

    var ndImageDiv = document.createElement("div");
    ndImageDiv.className = "image";
    ndImageDiv.id = "image-tag-" + image.id;
    ndImageDiv.onclick = setPrimaryImage.bind(window, image.id);

    var ndImage = document.createElement("img");
    ndImage.src = image.url + "_thumb.jpg";

    ndImageDiv.appendChild(ndImage);
    ndImageBlock.appendChild(ndImageDiv);
    ndImageBlock.appendChild(ndLink);
    thumbnailDiv.appendChild(ndImageBlock);
  }

  function setPrimaryImage(id)
  {
    var selected = document.getElementsByClassName("primaryImage");
    for(var i = 0, l = selected.length; i < l; i++)
    {
      selected[i].className = selected[i].className.replace(/primaryImage/g,
        "");
    }
    var toSelect = document.getElementById("image-tag-" + id);
    toSelect.className += " primaryImage";

    var field = document.getElementById("primaryImage");
    if(field != null)
    {
      field.value = id;
    }
  }

  function deleteImage(id)
  {
    new Dialog({
      title: "Are You Sure?",
      content: "Are you sure you would like to delete this image?",
      buttons: [
        {text: "YES", onclick: function()
          {
              $.ajax({
                url: "/webapi/image/delete/" + id,
                cache: false,
                data: {csrfToken: window.csrfToken},
                dataType: "json",
                method: "post",
                error: function()
                {
                  new Dialog({
                    title: "Failed To Delete Image",
                    content: "The image could not be deleted. Please " +
                      "refresh the page or try again later.",
                    buttons: [{text: "OK", onclick: function(){}}]
                  });
                }
              })
          }},
        {text: "NO", onclick: function(){}, isAlt: true}
      ]
    });
  }

  window.processNotification = function(msg)
  {
    var notif = msg.notification;
    var listing_id = document.getElementById("listing_id");
    if(msg.notif_type == "IM_PROCESS_DONE")
    {
      if(notif.image.media == "listing" &&
        notif.image.media_id == listing_id.value)
      {
        addImage(notif.image);
        removeProcessingBlock(notif.name);
      }
      return true;
    }
    else if(msg.notif_type == "IM_PROCESS_FAILED")
    {
      if(notif.media == "listing" && notif.media_id == listing_id.value)
      {
        removeProcessingBlock(notif.name);
      }
      return true;
    }
    else if(msg.notif_type == "IM_DELETE")
    {
      if(notif.media == "listing" && notif.media_id == listing_id.value)
      {
          var elem = document.getElementById("thumbnail-" + notif.id);
          if(elem)
          {
            imageCount--;
            elem.parentNode.removeChild(elem);
          }
      }
      return true;
    }
    return false;
  };

})( jQuery );
