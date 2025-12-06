import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import NotFound from "@/pages/NotFound";
import { Route, Switch } from "wouter";
import ErrorBoundary from "./components/ErrorBoundary";
import { ThemeProvider } from "./contexts/ThemeContext";
import Home from "./pages/Home";
import Dashboard from "./pages/Dashboard";
import Users from "./pages/Users";
import Offers from "./pages/Offers";
import Networks from "./pages/Networks";
import Teams from "./pages/Teams";
import Badges from "./pages/Badges";
import UserDetails from "./pages/UserDetails";
import Analytics from "./pages/Analytics";
// Phase 10: New Admin Pages
import Monitoring from "./pages/Monitoring";
import Tenants from "./pages/Tenants";
import GeoRules from "./pages/GeoRules";
import FraudDetection from "./pages/FraudDetection";
import LogsViewer from "./pages/LogsViewer";
import Webhooks from "./pages/Webhooks";
import Invoices from "./pages/Invoices";
import Contests from "./pages/Contests";

function Router() {
  // make sure to consider if you need authentication for certain routes
  return (
    <Switch>
      <Route path={"/"} component={Dashboard} />
      <Route path={"/users"} component={Users} />
      <Route path={"/users/:id"} component={UserDetails} />
      <Route path={"/offers"} component={Offers} />
      <Route path={"/networks"} component={Networks} />
      <Route path={"/teams"} component={Teams} />
      <Route path={"/badges"} component={Badges} />
      <Route path={"/analytics"} component={Analytics} />
      {/* Phase 10: System Pages */}
      <Route path={"/monitoring"} component={Monitoring} />
      <Route path={"/tenants"} component={Tenants} />
      <Route path={"/geo-rules"} component={GeoRules} />
      <Route path={"/fraud"} component={FraudDetection} />
      <Route path={"/logs"} component={LogsViewer} />
      <Route path={"/webhooks"} component={Webhooks} />
      <Route path={"/invoices"} component={Invoices} />
      <Route path={"/contests"} component={Contests} />
      <Route path={"/404"} component={NotFound} />
      {/* Final fallback route */}
      <Route component={NotFound} />
    </Switch>
  );
}

// NOTE: About Theme
// - First choose a default theme according to your design style (dark or light bg), than change color palette in index.css
//   to keep consistent foreground/background color across components
// - If you want to make theme switchable, pass `switchable` ThemeProvider and use `useTheme` hook

function App() {
  return (
    <ErrorBoundary>
      <ThemeProvider
        defaultTheme="dark"
      >
        <TooltipProvider>
          <Toaster />
          <Router />
        </TooltipProvider>
      </ThemeProvider>
    </ErrorBoundary>
  );
}

export default App;
