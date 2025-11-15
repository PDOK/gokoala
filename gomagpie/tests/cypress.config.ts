import { defineConfig } from "cypress";
import htmlvalidate from "cypress-html-validate/plugin";

export default defineConfig({
  e2e: {
    baseUrl: 'http://localhost:8080',
    setupNodeEvents(on, config) {
      htmlvalidate.install(on, {
        rules: {
          "require-sri": "off",
          "element-permitted-content": "off" // only because we use RDFa breadcrumbs
        },
      });
    },
  },
});
