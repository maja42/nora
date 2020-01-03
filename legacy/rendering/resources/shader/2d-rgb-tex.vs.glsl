uniform   mat4 vpMatrix;
uniform   mat4 modelTransform;

attribute vec2 position;
attribute vec3 color;
attribute vec2 texCoord;

varying vec3 vColor;
varying vec2 vTexCoord;

void main(void) {
    vec4 modelSpace      = vec4(position, 0.0, 1.0);
    vec4 worldSpace      = modelTransform * modelSpace;
    vec4 projectionSpace = vpMatrix * worldSpace;

    gl_Position = projectionSpace;
    vColor = color;
    vTexCoord = texCoord;
}
