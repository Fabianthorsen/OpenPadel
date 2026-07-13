import Root from './drawer.svelte';
import Trigger from './drawer-trigger.svelte';
import Content from './drawer-content.svelte';
import Header from './drawer-header.svelte';
import Title from './drawer-title.svelte';
import Description from './drawer-description.svelte';
import Body from './drawer-body.svelte';
import Footer from './drawer-footer.svelte';
import Close from './drawer-close.svelte';
import Overlay from './drawer-overlay.svelte';
import Portal from './drawer-portal.svelte';

export {
	drawerContentVariants,
	type DrawerPosition,
	type DrawerSize,
	type DrawerContentProps
} from './drawer-content.svelte';
export { type DrawerProps } from './drawer.svelte';
export { type DrawerTriggerProps } from './drawer-trigger.svelte';

export {
	Root,
	Trigger,
	Content,
	Header,
	Title,
	Description,
	Body,
	Footer,
	Close,
	Overlay,
	Portal,
	//
	Root as Drawer,
	Trigger as DrawerTrigger,
	Content as DrawerContent,
	Header as DrawerHeader,
	Title as DrawerTitle,
	Description as DrawerDescription,
	Body as DrawerBody,
	Footer as DrawerFooter,
	Close as DrawerClose,
	Overlay as DrawerOverlay,
	Portal as DrawerPortal
};
