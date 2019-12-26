var modeApp = new Vue({
  el: '#app',
  data: {
    mode: "minpv",
    gridPower: 0,
    pvPower: 0,
    chargePower: 0,
    socCharge: 0,
  },
  computed: {
  },
  methods: {
    activeMode: function (mode) {
      return mode == this.mode;
    },
    setMode: function (val) {
      axios.post('mode/' + val, {})
        .then(function (response) {
          mode = val;
          console.log(response);
        })
        .catch(function (error) {
          console.error(error);
        });
    },
    format: function (val) {
      return (Math.abs(val) >= 1e3) ? (val / 1e3).toFixed(2) : val.toFixed(0);
    },
    unit: function (val) {
      return (Math.abs(val) >= 1e3) ? "k" : "";
    },
    update: function (msg) {
      Object.keys(msg).forEach(function (k) {
        if (this[k] !== undefined) {
          this[k] = msg[k];
        }
      }, this);
    },
  },
  created: function() {
    const loc = window.location;
    const uri = loc.protocol + "//" + loc.hostname + (loc.port ? ":" + loc.port : "") + "/api";

    axios.defaults.baseURL = uri;
    axios.defaults.headers.post['Content-Type'] = 'application/json';
  }
})

$().ready(function () {
  connectSocket();
});

function connectSocket() {
  var ws, loc = window.location;
  var protocol = loc.protocol == "https:" ? "wss:" : "ws:"
  var uri = protocol + "//" + loc.hostname + (loc.port ? ":" + loc.port : "") + "/ws";

  ws = new WebSocket(uri);
  ws.onerror = function (evt) {
    ws.close();
  }
  ws.onclose = function (evt) {
    window.setTimeout(connectSocket, 500);
  };
  ws.onmessage = function (evt) {
    var msg = JSON.parse(evt.data);
    modeApp.update(msg);
  };
}
