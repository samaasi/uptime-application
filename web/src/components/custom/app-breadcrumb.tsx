import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { ProjectEnvironment } from '@/components/custom/project-environment';

export const AppBreadcrumb = () => {

  return (
    <div>
        <ProjectEnvironment />
    </div>
    // <Breadcrumb>
    //     <BreadcrumbList>
    //         <BreadcrumbItem className="hidden md:block">
    //             <BreadcrumbLink href="#">
    //               Building Your Application
    //             </BreadcrumbLink>
    //         </BreadcrumbItem>
    //         <BreadcrumbSeparator className="hidden md:block" />
    //         <BreadcrumbItem>
    //             <BreadcrumbPage>Data Fetching</BreadcrumbPage>
    //         </BreadcrumbItem>
    //     </BreadcrumbList>
    // </Breadcrumb>
        
  )
}