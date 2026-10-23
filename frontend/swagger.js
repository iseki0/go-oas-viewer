import SwaggerUI from 'swagger-ui-dist/swagger-ui-bundle.js';
import 'swagger-ui-dist/swagger-ui.css';
import { setup, showError } from './shared.js';

setup('swagger').then(url => {
  SwaggerUI({
    url,
    dom_id: '#app',
    deepLinking: true,
    withCredentials: true,
    persistAuthorization: false,
    validatorUrl: null,
  });
}).catch(showError);
