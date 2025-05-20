window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "/api/v1/openapi.yaml",
    dom_id: '#swagger-ui',
    presets: [SwaggerUIBundle.presets.apis],
    layout: "BaseLayout"
  });
};
