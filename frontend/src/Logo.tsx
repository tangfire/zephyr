type ZephyrLogoProps = {
  className?: string;
  title?: string;
};

export function ZephyrLogo({ className = "", title }: ZephyrLogoProps) {
  return (
    <svg
      className={`zephyr-logo ${className}`.trim()}
      viewBox="0 0 64 64"
      role={title ? "img" : undefined}
      aria-hidden={title ? undefined : true}
      focusable="false"
    >
      {title && <title>{title}</title>}
      <circle className="zephyr-logo-core" cx="32" cy="32" r="27" />
      <path className="zephyr-logo-wind" d="M18 20h26c2.8 0 4.8 2.2 4.3 4.7-.2 1.2-.9 2.2-1.8 3L25 45h26" />
      <path className="zephyr-logo-flare" d="M39.5 15.8 25.2 34h10.6l-5.2 14.2L45 29.4H34.2l5.3-13.6Z" />
      <circle className="zephyr-logo-node" cx="18" cy="20" r="3.6" />
      <circle className="zephyr-logo-node" cx="51" cy="45" r="3.6" />
      <path className="zephyr-logo-spark" d="M17 40c4.1 1.7 8.5 1.8 13.2.2" />
    </svg>
  );
}
