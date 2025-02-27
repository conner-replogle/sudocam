import { Navigate, Route, Routes } from "react-router";
import Layout from "./Layout";
import HomePage from "./pages/(dashboard)/HomePage";
import Auth from "./pages/Auth";
import DashLayout from "./pages/(dashboard)/Layout";
import AddCamera from "./pages/(dashboard)/cameras/add";

function App() {
  return (
    
      <Routes>

      
        <Route path="/" element={<Layout />}>

          <Route path="auth" element={<Auth />} />
          <Route path="dash" element={<DashLayout/>} >
            <Route path="cameras" >
              <Route path="add" element={<AddCamera />} />
            </Route>
            <Route index element={<HomePage />} />
          </Route>
{/* 
          <Route path="/create" element={<CreatePage />} />
          <Route path="/auth" element={<AuthPage />} /> */}
          {/* <Route path="pgp" element={<PgpHomePage/>}/>
          <Route path="pgp/generate" element={<Generate />} />
          <Route path="pgp/message" element={<Message />} />
          <Route path="pgp/identities" element={<Identities />} /> */}


          {/* Using path="*"" means "match anything", so this route
                acts like a catch-all for URLs that we don't have explicit
                routes for. */}
          <Route path="*" element={<NoMatch />} />
        </Route>
      </Routes>
      
      
  );
}

function NoMatch(){
  return <div>
    <Navigate to="/" />
    <h1>Not FOund</h1>
    
    </div>
}

export default App;