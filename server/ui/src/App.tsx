import { Navigate, Route, Routes } from "react-router";
import Layout from "./Layout";
import HomePage from "./pages/(dashboard)/HomePage";
import Auth from "./pages/Auth";
import DashLayout from "./pages/(dashboard)/Layout";
import AddCamera from "./pages/(dashboard)/cameras/add";
import { CameraPage } from "./pages/(dashboard)/cameras/camera";
import { GroupCameras } from "./pages/(dashboard)/groups/groups";
import { CamerasListPage } from "./pages/(dashboard)/cameras";
import { RecordedPage } from "./pages/(dashboard)/record/RecordedPage";
import { UserProvider } from "./context/UserContext";

function App() {
  return (
      <UserProvider>
      <Routes>

          <Route element={<Layout />} >
            <Route path="auth" element={<Auth />} />
          </Route>
          <Route  element={<DashLayout/>} >
            <Route path="recordings" >
              <Route index element={<CamerasListPage />} />

              <Route path=":id" element={<RecordedPage />} />
            </Route>
            <Route path="cameras" >
              <Route path="add" element={<AddCamera />} />
              <Route path=":id" element={<CameraPage />} />
              <Route index element={<CamerasListPage />} />
            </Route>
            
            <Route path="groups" element={<HomePage />}>
              <Route path=":group_id" element={<GroupCameras />} />

            </Route>
            <Route index element={<HomePage />} />

          </Route>


        <Route path="*" element={<NoMatch />} />
      </Routes>
      </UserProvider>
      
      
  );
}

function NoMatch(){
  return <div>404 Not Found</div>
}

export default App;