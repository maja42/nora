precision mediump float;

uniform sampler2D sampler;
uniform vec4 color;

varying vec2 vTexCoord;

void main(void) {
    // go textures have their origin in the top-left corner.
    // openGL expects it in the bottom-left corner.
    // Therefore, we need to flip the texture vertically.
    vec4 texel = texture2D(sampler, vec2(vTexCoord.s, -vTexCoord.t));
    texel = texel * color;

    gl_FragColor = texel;
}