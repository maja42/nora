uniform   mat4 vpMatrix;            // contains view + projection matrix
uniform   mat4 modelTransform;

attribute vec2 position;
attribute vec2 texCoord;

varying vec2 vTexCoord;

void main(void) {
    vec4 modelSpace      = vec4(position, 0.0, 1.0);
    vec4 worldSpace      = modelTransform * modelSpace;
    vec4 projectionSpace = vpMatrix * worldSpace;

    gl_Position = projectionSpace;
    vTexCoord = texCoord;
}
