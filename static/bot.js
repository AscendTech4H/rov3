//this is dumb

var bleh = document.getElementById("bleh");
//inject preloader
bleh.innerHTML = `
<div class="preloader-wrapper big active">
    <div class="spinner-layer spinner-blue-only">
      <div class="circle-clipper left">
        <div class="circle"></div>
      </div><div class="gap-patch">
        <div class="circle"></div>
      </div><div class="circle-clipper right">
        <div class="circle"></div>
      </div>
    </div>
  </div>
`;
//metadata from info.json
var meta;

var camn = 0;   //number of connected cameras
var qn = 0; //query breaker (incremented every frame)
var player; //video player div

function promReq(url) {
    return new Promise((s, f) => {
        var xhr = new XMLHttpRequest();
        xhr.addEventListener("load", () => {
            s(xhr.responseText);
        });
        xhr.addEventListener("error", () => {f(xhr.statusText);});
        xhr.open("GET", url);
        xhr.send();
    })
}
function jsonReq(url) {
    return new Promise((s, f) => {
        promReq(url).then((dat) => {
            var o;
            try {
                o = JSON.parse(dat);
            } catch(err) {
                f(err);
                return;
            }
            s(o);
        }, f);
    })
}
function promImg(url) { //load an image from a URL asynchronously
    return new Promise((s, f) => {
        var img = new Image();
        img.onload = () => {
            s(img);
        };
        img.onerror = () => {
            f();
        };
        img.src = url;
    })
}
//load a materialize icon
function genIcon(name) {
    var icon = document.createElement("i");
    icon.classList.add("materialize-icons");
    icon.innerHTML = name;
    return icon;
}
//make a materialize button
function genButton(elems, action) {
    var btn = document.createElement("a");          //create <a>
    btn.classList.add("btn");                       //its a button
    btn.classList.add("waves-effect");              //make a wave effect when clicked
    btn.classList.add("waves-light");               //light waves
    elems.forEach((el) => {btn.appendChild(el);});  //add contents
    btn.onclick = action;                           //set click handler
    return btn;
}
//buttons for video player
var btns = (() => {
    var div = document.createElement("div");
    div.classList.add("row");
    div.appendChild(
        genButton(
            [
                genIcon("navigate_before"),
                document.createTextNode("cam")
            ],
            () => {
                if(camn == 0) {
                    camn = meta.NCameras;
                }
                camn--;
            }
        )
    );
    div.appendChild(
        genButton(
            [
                document.createTextNode("cam"),
                genIcon("navigate_next")
            ],
            () => {
                camn++;
                if(camn == meta.NCameras) {
                    camn = 0;
                }
            }
        )
    );
    return div;
})();
//current video frame image object
var frame = document.createElement("img");   //element containing video frame
var fps = document.createElement("i");        //fps element
//last frame update time
var lastupdate = Date.now();
function updateVideo(imgu) { //update the video player display
    //update FPS counter
    var now = Date.now();
    fps.innerHTML = 'FPS: ' + (1000.0/(now - lastupdate));
    lastupdate = now;
    //switch to new frame
    frame.src = imgu;
}
player = bleh;
var isloading = false;
var isrunok = false;
function runFrame() {
    qn++;
    updateVideo("/bs/"+camn+"/frame.jpg?qn="+qn);
    isloading = true;
    isrunok = false;
}
function handleLoad() {
    isloading = false;
    if(isrunok) {
        runFrame();
    }
}
function handleTime() {
    isrunok = true;
    if(!isloading) {
        runFrame();
    }
}
function startPlayer() {
    bleh.innerHTML = '';
    bleh.appendChild(frame);
    bleh.appendChild(fps);
    bleh.appendChild(btns);
    frame.onload = handleLoad;
    setInterval(handleTime, 1000/30);
    frame.onerr = () => {
        bleh.innerHTML = "failed to load frame";
    }
    runFrame();
}

//start it all
jsonReq('/bs/info.json').then(
    (dat) => {meta = dat;startPlayer();},
    (err) => {bleh.innerHTML = document.createTextNode(err).outerHTML;}
)
