package io.cellery.models;

import lombok.Data;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Cell/Composite Image model Class.
 */
@Data
public class Image {
    private Map<String, Component> componentNameToComponentMap;
    private String orgName;
    private String cellName;
    private String cellVersion;
    private List<Dependency> dependencies;
    private boolean zeroScaling;
    private boolean autoScaling;
    private boolean compositeImage;
    private Test test;

    public Image() {
        componentNameToComponentMap = new HashMap<>();
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

    public void addDependency(Dependency dependency) {
        dependencies.add(dependency);
    }
}
