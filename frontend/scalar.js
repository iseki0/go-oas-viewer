import { createApiReference } from '@scalar/api-reference';
import '@scalar/api-reference/style.css';
import { setup, showError } from './shared.js';

setup('scalar').then(url => {
  createApiReference('#app', {
    url,
    withDefaultFonts: false,
    proxyUrl: '',
    customFetch: (input, init) => fetch(input, { ...init, credentials: 'include' }),
  });
}).catch(showError);
