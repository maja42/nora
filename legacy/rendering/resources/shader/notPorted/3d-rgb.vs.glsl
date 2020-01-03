uniform   mat4 vpMatrix;
uniform   mat4 modelTransform;

attribute vec3 position;
attribute vec3 color;

varying vec3 vColor;

void main(void) {
    vec4 modelSpace      = vec4(position, 1.0);
    vec4 worldSpace      = modelTransform * modelSpace;
    vec4 projectionSpace = vpMatrix * worldSpace;

    gl_Position = projectionSpace;
    vColor = color;
}
