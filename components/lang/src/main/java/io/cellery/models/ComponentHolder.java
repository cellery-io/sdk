package io.cellery.models;

import org.ballerinalang.util.exceptions.BallerinaException;

import java.util.HashMap;
import java.util.Map;

/**
 * Component Holder Class.
 */
public class ComponentHolder {

    private Map<String, Component> componentNameToComponentMap;

    public ComponentHolder() {
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
