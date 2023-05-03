import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import App from './App';
import { BrowserRouter } from "react-router-dom";
import { CookiesProvider }  from 'react-cookie';
import axios from 'axios';

axios.defaults.baseURL = window.ROUTE_PREFIX + "/api/v1/";

ReactDOM.render(
  <React.StrictMode>
    <BrowserRouter basename={window.ROUTE_PREFIX}>
      <CookiesProvider>
        <App />
      </CookiesProvider>
    </BrowserRouter>
  </React.StrictMode>,
  document.getElementById('root')
);
