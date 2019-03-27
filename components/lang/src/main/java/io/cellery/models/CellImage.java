package io.cellery.models;

import lombok.Data;
import org.ballerinalang.util.exceptions.BallerinaException;

import java.util.HashMap;
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

    public CellImage() {
        componentNameToComponentMap = new HashMap<>();
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

    public Component getComponent(String componentName) {
        Component temp = componentNameToComponentMap.get(componentName);
        if (temp == null) {
            throw new BallerinaException("Invalid component name " + componentName);
        }
        return temp;
    }
}
