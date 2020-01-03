uniform   mat4 vpMatrix;
uniform   mat4 modelTransform;

attribute vec3 position;
attribute vec3 normal;

uniform vec4 fragColor;

varying vec4 vColor;


struct DirectionalLight {
    vec3  color;
    vec3  direction;
    float ambientIntensity;
    float diffuseIntensity;
};

vec4 CalcLight(DirectionalLight light, vec3 normal) {
    vec4 ambientColor =   vec4(light.color * light.ambientIntensity, 1.0);
    float diffuseFactor = dot(normal, -light.direction);

    vec4 diffuseColor  = vec4(0, 0, 0, 0);

    if (diffuseFactor > 0.0) {
        diffuseColor = vec4(light.color * light.diffuseIntensity * diffuseFactor, 1.0);
    }
    return ambientColor + diffuseColor;
}

void main(void) {
    vec4 modelSpace      = vec4(position, 1.0);
    vec4 worldSpace      = modelTransform * modelSpace;
    vec4 projectionSpace = vpMatrix * worldSpace;
    
    gl_Position = projectionSpace;
 
    vec3 worldSpaceNormal = mat3(modelTransform) * normal;
    worldSpaceNormal      = normalize(worldSpaceNormal);
    // This simplified normal conversion is allowed for uniform-scaling only.
    // Otherwise, inverse and transpose matrices must be determined.

    DirectionalLight light = DirectionalLight(
        vec3( 1, 1, 1), // light color
        vec3(-1,-1,+1), // direction
        0.3,            // ambient intensity
        0.7             // diffuse intensity
    );

    // gouraud shading is good enough - no phong shading required.
    // (The objects rendered with this shader are expected to have no curved surfaces)
    vec4 totalLight = CalcLight(light, worldSpaceNormal);
    vColor = fragColor * totalLight;
}
