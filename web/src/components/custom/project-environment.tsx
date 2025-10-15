'use client';

import * as React from 'react';
import Image from 'next/image';
import { cn } from '@/lib/utils';
import {
    Select,
    SelectItem,
    SelectValue,
    SelectContent,
    SelectTrigger,
} from '@/components/ui/select';
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbList,
    BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import { ChevronsUpDown } from 'lucide-react';

type EnvironmentStatus = 'development' | 'staging' | 'production';

interface Environment {
  value: string;
  label: string;
  status: EnvironmentStatus;
}

interface Project {
  value: string;
  label: string;
  image: string;
  environments: Environment[];
}

const Projects: Project[] = [
  {
    value: 'proj_1',
    label: 'Main Project',
    image: 'https://placehold.co/20x20/7c3aed/ffffff',
    environments: [
      { value: 'proj_1-dev', label: 'Development', status: 'development' },
      { value: 'proj_1-prod', label: 'Production', status: 'production' },
    ],
  },
  {
    value: 'proj_2',
    label: 'Origin Project',
    image: 'https://placehold.co/20x20/ea580c/ffffff',
    environments: [
      { value: 'proj_2-staging', label: 'Staging', status: 'staging' },
      { value: 'proj_2-prod', label: 'Production', status: 'production' },
    ],
  },
];

const StatusIndicator = React.memo(({ status }: { status: EnvironmentStatus }) => {
  const statusColorMap: Record<EnvironmentStatus, string> = {
    development: 'bg-amber-700',
    staging: 'bg-gray-500',
    production: 'bg-green-500',
  };

  return (
    <div
      className={cn('w-2 h-2 rounded-full', statusColorMap[status] || 'bg-gray-400')}
      role="presentation"
      title={`Status: ${status}`}
    />
  );
});
StatusIndicator.displayName = 'StatusIndicator';

export const ProjectEnvironment = () => {
  const [selectedProjectValue, setSelectedProjectValue] = React.useState(Projects[0].value);
  
  const selectedProject = React.useMemo(
    () => Projects.find((p) => p.value === selectedProjectValue) || Projects[0],
    [selectedProjectValue]
  );
  
  const [selectedEnvValue, setSelectedEnvValue] = React.useState(selectedProject.environments[0].value);

  const selectedEnvironment = React.useMemo(
    () =>
      selectedProject.environments.find((e) => e.value === selectedEnvValue) ||
      selectedProject.environments[0],
    [selectedEnvValue, selectedProject.environments]
  );

  const handleProjectChange = (newProjectValue: string) => {
    const newProject = Projects.find((p) => p.value === newProjectValue);
    if (newProject) {
      setSelectedProjectValue(newProjectValue);
      setSelectedEnvValue(newProject.environments[0]?.value || '');
    }
  };

  return (
    <div className="flex items-center justify-between bg-gray-100 dark:bg-gray-800 shadow-md rounded-sm p-1 -m-px">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <Select defaultValue={selectedProjectValue} onValueChange={handleProjectChange}>
              <CustomSelectTrigger>
                <SelectValue>
                  <div className="flex items-center gap-2">
                    <Image
                      width={16}
                      height={16}
                      className="rounded-sm object-cover"
                      src={selectedProject.image}
                      alt={`${selectedProject.label} logo`}
                    />
                    <span>{selectedProject.label}</span>
                  </div>
                </SelectValue>
              </CustomSelectTrigger>
              <SelectContent>
                {Projects.map((project) => (
                  <SelectItem key={project.value} value={project.value}>
                    <div className="flex items-center gap-2">
                      <Image
                        src={project.image}
                        alt={`${project.label} logo`}
                        className="w-4 h-4 rounded-sm object-cover"
                        width={16}
                        height={16}
                      />
                      <span>{project.label}</span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </BreadcrumbItem>
          <BreadcrumbSeparator> / </BreadcrumbSeparator>
          <BreadcrumbItem>
            <Select defaultValue={selectedEnvValue} onValueChange={setSelectedEnvValue}>
              <CustomSelectTrigger>
                <SelectValue>
                  <div className="flex items-center gap-2">
                    <StatusIndicator status={selectedEnvironment.status} />
                    <span>{selectedEnvironment.label}</span>
                  </div>
                </SelectValue>
              </CustomSelectTrigger>
              <SelectContent>
                {selectedProject.environments.map((env) => (
                  <SelectItem key={env.value} value={env.value}>
                    <div className="flex items-center gap-2">
                      <StatusIndicator status={env.status} />
                      <span>{env.label}</span>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </div>
  );
};

const CustomSelectTrigger = ({ children }: { children: React.ReactNode }) => (
  <SelectTrigger className="focus-visible:bg-accent h-8 px-1.5 focus-visible:ring-0 focus-visible:ring-offset-0 border-none shadow-none hover:shadow-sm bg-transparent hover:bg-accent [&>svg:not(.lucide-chevrons-up-down)]:hidden">
    {children}
    <ChevronsUpDown size={14} className="text-muted-foreground/80 ml-1 lucide-chevrons-up-down" />
  </SelectTrigger>
);