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

    public void addAPI(String componentName, API api) {
        Component temp = componentNameToComponentMap.remove(componentName);
        if (temp == null) {
            throw new BallerinaException("Invalid component name " + componentName);
        }
        api.setBackend(temp.getService());
        temp.addApi(api);
        componentNameToComponentMap.put(componentName, temp);
    }

    public void addTCP(String componentName, TCP tcp) {
        Component temp = componentNameToComponentMap.remove(componentName);
        if (temp == null) {
            throw new BallerinaException("Invalid component name " + componentName);
        }
        tcp.setBackendHost(temp.getService());
        temp.addTCP(tcp);
        componentNameToComponentMap.put(componentName, temp);
    }
}
