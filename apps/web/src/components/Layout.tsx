import { motion } from 'framer-motion'
import { Link, Outlet } from 'react-router-dom'

export function Layout() {
  return <div className='app'><motion.aside initial={{x:-20, opacity:0}} animate={{x:0, opacity:1}} className='sidebar'><h3>Logs</h3><Link to='/traces'>Traces</Link><a>Settings</a></motion.aside><main><Outlet/></main></div>
}
