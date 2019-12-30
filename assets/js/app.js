// use for testing purposes
const baseurl = {
  protocol: "http:",
  hostname: "localhost",
  port: "7070",
};

const mode = new Vue({
  el: '#mode',
  data: {
    mode: null,
    error: null,
  },
  computed: {
    modeOff: function() { return this.mode == "off"; },
    modeNow: function() { return this.mode == "now"; },
    modeMinPV: function() { return this.mode == "minpv"; },
    modePV: function() { return this.mode == "pv"; },
  },
  methods: {
    setMode: function (val) {
      self = this;
      axios.post('mode/' + val, {}).then(function (response) {
        self.mode = response.data.mode;
      });
    },
  },
  created: function() {
    const loc = baseurl || window.location;
    const uri = loc.protocol + "//" + loc.hostname + (loc.port ? ":" + loc.port : "") + "/api";

    axios.defaults.baseURL = uri;
    axios.defaults.headers.post['Content-Type'] = 'application/json';

    // error handler
    const self = this;
    axios.interceptors.response.use(function (response) {
      return response;
    }, function (error) {
      self.error = error;
      window.setTimeout(function () {
        if (self.error == error) { self.error = ""; }
      }, 5000);
      return Promise.reject(error);
    });
  },
  mounted: function() {
    const self = this;
    axios.get('mode').then(function (response) {
      self.mode = response.data.mode;
    });
  },
});

const live = new Vue({
  el: '#live',
  data: {
    gridPower: 0,
    pvPower: 0,
    chargeCurrent: 0,
    chargePower: 0,
    chargeEnergy: 0,
    socCharge: 0,
  },
  computed: {
    gridMode: function () {
      return (this.gridPower >= 0) ? "Bezug" : "Einspeisung";
    },
  },
  methods: {
    format: function (val) {
      val = Math.abs(val);
      return (val >= 1e3) ? (val / 1e3).toFixed(1) : val.toFixed(0);
    },
    unit: function (val) {
      return (Math.abs(val) >= 1e3) ? "k" : "";
    },
    update: function (msg) {
      Object.keys(msg).forEach(function (k) {
        if (this[k] !== undefined) {
          this[k] = msg[k];
        } else if (mode[k] !== undefined) {
          mode[k] = msg[k]; // send to mode app
        } else {
          console.error("invalid data key: " + k)
        }
      }, this);
    },
    connect: function () {
      const loc = baseurl || window.location;
      const protocol = loc.protocol == "https:" ? "wss:" : "ws:"
      const uri = protocol + "//" + loc.hostname + (loc.port ? ":" + loc.port : "") + "/ws";
      const ws = new WebSocket(uri), self = this;
      ws.onerror = function (evt) {
        ws.close();
      };
      ws.onclose = function (evt) {
        window.setTimeout(self.connect, 1000);
      };
      ws.onmessage = function (evt) {
        var msg = JSON.parse(evt.data);
        console.log(msg)
        self.update(msg);
      };
    },
  },
  created: function () {
    this.connect();
  },
});
