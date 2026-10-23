import '@stoplight/elements/web-components.min.js';
import '@stoplight/elements/styles.min.css';
import { setup, showError } from './shared.js';

setup('elements').then(async url => {
  await customElements.whenDefined('elements-api');
  const viewer = document.createElement('elements-api');
  viewer.setAttribute('router', 'hash');
  viewer.setAttribute('layout', 'responsive');
  viewer.apiDescriptionUrl = url;
  viewer.tryItCredentialsPolicy = 'include';
  viewer.tryItCorsProxy = '';
  document.querySelector('#app').append(viewer);
}).catch(showError);
