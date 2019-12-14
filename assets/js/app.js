var modeApp = new Vue({
  el: '#app',
  data: {
    mode: "now",
  },
  computed: {
    modeSelectedNow: function() {
      return this.mode == "now"
    },
    modeSelectedMinPV: function() {
      return this.mode == "minpv"
    },
    modeSelectedPV: function() {
      return this.mode == "pv"
    },
  }
})

// $.ajax('myservice/username', {
//     data: {
//       id: 'some-unique-id'
//     }
//   })
//   .then(
//     function success(name) {
//       alert('User\'s name is ' + name);
//     },
//     function fail(data, status) {
//       alert('Request failed.  Returned status of ' + status);
//     }
//   );