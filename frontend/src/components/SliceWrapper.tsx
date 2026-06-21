import React, { useEffect, useRef } from 'react';

interface SliceWrapperProps {
  html: string;
  ControllerClass: any;
  context: {
    active_selected_legal_entity_id: string;
    gatewayUrl?: string;
    token?: string;
  };
}

export const SliceWrapper: React.FC<SliceWrapperProps> = ({ html, ControllerClass, context }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const controllerRef = useRef<any>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    // Render HTML markup inside the viewport container
    containerRef.current.innerHTML = html;

    // Instantiate the Vanilla JS controller and initialize it
    const controller = new ControllerClass(containerRef.current, context);
    controllerRef.current = controller;
    controller.init();

    return () => {
      // Destructor to clean up event listeners/outbox subscriptions
      if (controllerRef.current && typeof controllerRef.current.destroy === 'function') {
        controllerRef.current.destroy();
      }
    };
  }, [html, ControllerClass, context]);

  return <div ref={containerRef} className="w-full h-full" />;
};
