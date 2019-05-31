package io.cellery.models;

import lombok.Data;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Cell Image model Class.
 */
@Data
public class CellImage {
    private Map<String, Component> componentNameToComponentMap;
    private String orgName;
    private String cellName;
    private String cellVersion;
    private List<String> dockerImages;
    private List<Dependency> dependencies;

    public CellImage() {
        componentNameToComponentMap = new HashMap<>();
        dockerImages = new ArrayList<>();
        dependencies = new ArrayList<>();
    }

    public Map<String, Component> getComponentNameToComponentMap() {
        return componentNameToComponentMap;
    }

    public void addComponent(Component component) {
        if (componentNameToComponentMap.containsKey(component.getName())) {
            throw new BallerinaException("Two components with same name exists " + component.getName());
        }
        this.componentNameToComponentMap.put(component.getName(), component);
    }

    public void addDockerImage(String imageName) {
        dockerImages.add(imageName);
    }

    public void addDependency(Dependency dependency) {
        dependencies.add(dependency);
    }
}
