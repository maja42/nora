uniform   mat4 vpMatrix;            // contains view + projection matrix
uniform   mat4 modelTransform;

attribute vec3 position;
attribute vec2 texCoord;

varying vec2 vTextureCoord;

void main(void) {
    vec4 modelSpace      = vec4(position, 1.0);                    // homogenous 3D space
    vec4 worldSpace      = modelTransform * modelSpace;
    vec4 projectionSpace = vpMatrix * worldSpace;

    gl_Position = projectionSpace;
    vTextureCoord = texCoord;
}
