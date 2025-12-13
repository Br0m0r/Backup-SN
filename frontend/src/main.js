import { createApp } from 'vue';
import App from './App.vue';
import './assets/base.css';
import router from './router';
import { restoreSession } from './stores/auth';
import './services/axios'; // Import axios interceptor

// Restore user session from localStorage before mounting app
restoreSession();

const app = createApp(App);
app.use(router);
app.mount('#app');
