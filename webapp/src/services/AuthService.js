import axios from 'axios';


export default {
  login(token) {
    return axios
      .get('/list',{params: {token: token}})
      .then(response => response.data);
  },
}